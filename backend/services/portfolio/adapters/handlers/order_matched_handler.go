package handler_adapters

import (
	"brokerx/portfolio-service/models"
	"brokerx/portfolio-service/ports"
	"brokerx/portfolio-service/util"
	"context"
)

type OrderMatchedHandler struct {
	Producer ports.EventProducer
	Tm 		 ports.TransactionManager
}

func (h *OrderMatchedHandler) handle(ctx context.Context, event models.MatchingEvent) error {
	log := util.FromContext(ctx)

	if len(event.Executions) == 0 {
		return h.successEvent(ctx, event)
	}

	order := event.Order
	err := h.Tm.Do(ctx, func(or ports.OrderRepository, er ports.ExecutionRepository, wr ports.WalletRepository, pr ports.PositionRepository, obr ports.OutboxRepository) error {
		total, qty := getFundsNeeded(event.Executions)
		if err := updateWallets(ctx, order, total, event.Orders, wr); err != nil {
			log.Error("error when updating wallets", "error", err)
			if err := revertPositionsReservations(ctx, pr, order, qty, event.Orders); err != nil {
				log.Error("error reverting positions reservations", "error", err)
				return err
			}
			return revertOrdersAndCreateOutboxEvents(ctx, obr, event, qty, err) // this commits or rolls back the transaction for the failure path
		}

		if err := updatePositionsQuantities(ctx, order, qty, event.Orders, pr); err != nil {
			log.Error("error when updating positions", "error", err)
			return err
		}

		if err := er.CreateBatch(ctx, event.Executions); err != nil {
			log.Error("execution batch failed", "error", err)
			return err
		}

		var ordersToUpdate = []*models.Order{&order}
		for _, o := range event.Orders {
			ordersToUpdate = append(ordersToUpdate, &o.Order)
		}
		if err := or.UpdateBatch(ctx, ordersToUpdate); err != nil {
			log.Error("order batch update failed", "error", err)
			return err
		}

		if err := createSuccessOutboxEvents(ctx, obr, event); err != nil {
			log.Error("error creating success outbox events", "error", err)
			return err
		}

		return nil // this commits the transaction for the successful path
	})

	if err != nil {
		log.Error("error confirming order matching. OrderMatched event will be consumed until successful...", "orderId", order.ID, "error", err)
		return err
	}

	// Dont send any event. Outbox dispatcher will send events asynchronously.
	return nil
}

func (h *OrderMatchedHandler) successEvent(ctx context.Context, event models.MatchingEvent) error {
	orderEvent := createOrderEvent(event, &event.Order)
	return h.Producer.SendEvent(ctx, orderEvent.Topic, orderEvent.Event, orderEvent.Order, nil)
}

func updateWallets(ctx context.Context, order models.Order, total float64, claimedOrders []*models.ClaimedCandidate, walletRepo ports.WalletRepository) error {
	deltas := []models.WalletDelta{{Order: order, Total: total}}
	for _, co := range claimedOrders {
		deltas = append(deltas, models.WalletDelta{Order: co.Order, Total: pickUnitPrice(&order, &co.Order) * float64(co.ClaimedQty)})
	}
	return walletRepo.ReleaseFunds(ctx, deltas)
}

func pickUnitPrice(incoming, candidate *models.Order) float64 {
	if incoming.Type == "market" || candidate.Type == "limit" {
		return candidate.UnitPrice
	}
	return incoming.UnitPrice
}

func updatePositionsQuantities(ctx context.Context, order models.Order, qty int, claimedOrders []*models.ClaimedCandidate, posRepo ports.PositionRepository) error {
	incomingClaimed := []*models.ClaimedCandidate{{Order: order, ClaimedQty: qty}}
	if order.Action == "sell" {
		if err := posRepo.ReleaseQuantity(ctx, incomingClaimed); err != nil {
			return err
		}
		if err := posRepo.AddAvailableQuantity(ctx, claimedOrders); err != nil {
			return err
		}
		return nil
	}

	if err := posRepo.AddAvailableQuantity(ctx, incomingClaimed); err != nil {
		return err
	}
	if err := posRepo.ReleaseQuantity(ctx, claimedOrders); err != nil {
		return err
	}
	return nil
}

func createSuccessOutboxEvents(ctx context.Context, obr ports.OutboxRepository, event models.MatchingEvent) error {
	successEvent := createOrderEvent(event, &event.Order)
	var orderEvents = []*models.OrderEvent{&successEvent}

	for _, o := range event.Orders {
		successEvent := createOrderEvent(event, &o.Order)
		orderEvents = append(orderEvents, &successEvent)
	}
	return obr.CreateOrderEvents(ctx, orderEvents)
}

func createOrderEvent(event models.MatchingEvent, order *models.Order) models.OrderEvent {
	topic := "MatchingEvents"
	eventType := "OrderOpen"
	if order.Status == "filled" {
		eventType = "OrderConfirmed"
		topic = "NotificationEvents"
	}
	return models.OrderEvent{
		Topic: topic,
		Event: eventType,
		TraceID: event.TraceID,
		JWT: event.JWT,
		Order: *order,
	}
}

func revertPositionsReservations(ctx context.Context, posRepo ports.PositionRepository, incomingOrder models.Order, qty int, claimedOrders []*models.ClaimedCandidate) error {
	deltas := []models.PositionDelta{{UserID: incomingOrder.UserID, Symbol: incomingOrder.Symbol, Qty: qty}}
	for _, co := range claimedOrders {
		deltas = append(deltas, models.PositionDelta{UserID: co.Order.UserID, Symbol: co.Order.Symbol, Qty: co.ClaimedQty})
	}
	return posRepo.RevertReservations(ctx, deltas)
}

func revertOrdersAndCreateOutboxEvents(ctx context.Context, obr ports.OutboxRepository, event models.MatchingEvent, qty int, err error) error {
	updateStatusAndQty(&event.Order, qty)
	var orderEvents = []*models.OrderEvent{{
		Topic: "OrderEvents",
		Event: "OrderConfirmationFailed",
		TraceID: event.TraceID,
		Order: event.Order,
		Error: err.Error(),
	}}

	for _, o := range event.Orders {
		updateStatusAndQty(&o.Order, o.ClaimedQty)
		returningEvent := createOrderEvent(event, &o.Order)
		orderEvents = append(orderEvents, &returningEvent)
	}

	return obr.CreateOrderEvents(ctx, orderEvents) // this commits the transaction for the failure path
}

func updateStatusAndQty(order *models.Order, qty int) {
	order.RemainingQuantity += qty
	if order.RemainingQuantity != 0 && order.RemainingQuantity < order.Quantity {
		order.Status = "partially_filled"
		return
	}

	if order.RemainingQuantity == 0 {
		order.Status = "filled"
		return
	}

	if order.RemainingQuantity == order.Quantity {
		order.Status = "open"
		return
	}
}

func getFundsNeeded(records []*models.ExecutionRecord) (float64, int) {
	total := 0.0
	qty := 0

	for _, r := range records {
		total += r.Price * float64(r.Quantity)
		qty += r.Quantity
	}

	return total, qty
}
