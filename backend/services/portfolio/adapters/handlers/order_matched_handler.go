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

	incoming := event.Order
	err := h.Tm.Do(ctx, func(er ports.ExecutionRepository, wr ports.WalletRepository, pr ports.PositionRepository, obr ports.OutboxRepository) error {
		originalQty := getQuantityNeeded(event.Executions)
		qty, err := updateWalletsAndAdjustOrders(ctx, &incoming, originalQty, &event.Orders, wr)
		if err != nil {
			log.Error("error when updating wallets", "error", err)
			if err := revertPositionsReservations(ctx, pr, incoming, originalQty, event.Orders); err != nil {
				log.Error("error reverting positions reservations", "error", err)
				return err
			}
			return revertOrdersAndCreateOutboxEvents(ctx, obr, event, originalQty, err) // this commits or rolls back the transaction for the failure path
		}

		if err := updatePositionsQuantities(ctx, incoming, qty, event.Orders, pr); err != nil {
			log.Error("error when updating positions", "error", err)
			return err
		}

		if err := er.CreateBatch(ctx, event.Executions); err != nil {
			log.Error("execution batch failed", "error", err)
			return err
		}

		if err := createSuccessOutboxEvents(ctx, obr, incoming, event); err != nil {
			log.Error("error creating success outbox events", "error", err)
			return err
		}

		return nil // this commits the transaction for the successful path
	})

	if err != nil {
		log.Error("error confirming order matching. OrderMatched event will be consumed until successful...", "orderId", incoming.ID, "error", err)
		return err
	}

	// Dont send any event. Outbox dispatcher will send events asynchronously.
	return nil
}

func (h *OrderMatchedHandler) successEvent(ctx context.Context, event models.MatchingEvent) error {
	orderEvent := createOrderEvent(event, &event.Order)
	return h.Producer.SendEvent(ctx, orderEvent.Topic, orderEvent.Event, orderEvent.Order, nil)
}

func updateWalletsAndAdjustOrders(ctx context.Context, incoming *models.Order, originalQty int, claimedOrders *[]*models.ClaimedCandidate, walletRepo ports.WalletRepository) (int, error) {
	deltas :=  make(map[int]models.WalletDelta)
	for _, co := range *claimedOrders {
		deltas[co.Order.ID] = models.WalletDelta{Order: co.Order, Total: pickUnitPrice(incoming, &co.Order) * float64(co.ClaimedQty)}
	}

	deltas, err := walletRepo.ReleaseFunds(ctx, deltas)
	if err != nil {
		return 0, err
	}

	totalToDeductFromIncoming := 0.0
	newQty := 0
	for _, co := range *claimedOrders {
		d := deltas[co.Order.ID]
		if d.Total == 0 {
			co.Order.Status = "canceled"
			co.Order.RemainingQuantity += co.ClaimedQty
			co.ClaimedQty = 0
			continue
		}
		newQty += co.ClaimedQty
		totalToDeductFromIncoming += d.Total
	}

	var incomingDelta models.WalletDelta
	incomingDelta.Order = *incoming
	incomingDelta.Total = totalToDeductFromIncoming
	_, err = walletRepo.ReleaseFunds(ctx, map[int]models.WalletDelta{incoming.ID: incomingDelta})

	incoming.RemainingQuantity += originalQty - newQty
	updateStatus(incoming)

	return newQty, err
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

func createSuccessOutboxEvents(ctx context.Context, obr ports.OutboxRepository, incoming models.Order, event models.MatchingEvent) error {
	successEvent := createOrderEvent(event, &incoming)
	var orderEvents = []*models.OrderEvent{&successEvent}

	for _, o := range event.Orders {
		successEvent := createOrderEvent(event, &o.Order)
		orderEvents = append(orderEvents, &successEvent)
	}
	return obr.CreateOrderEvents(ctx, orderEvents)
}

func createOrderEvent(event models.MatchingEvent, order *models.Order) models.OrderEvent {
	return models.OrderEvent{
		Topic: "OrderEvents",
		Event: "OrderConfirmed",
		TraceID: event.TraceID,
		JWT: event.JWT,
		Order: *order,
	}
}

func revertPositionsReservations(ctx context.Context, posRepo ports.PositionRepository, incomingOrder models.Order, qty int, claimedOrders []*models.ClaimedCandidate) error {
	var deltas []models.PositionDelta
	if incomingOrder.Action == "sell" {
		deltas = []models.PositionDelta{{UserID: incomingOrder.UserID, Symbol: incomingOrder.Symbol, Qty: qty}}
	} else {
		for _, co := range claimedOrders {
			deltas = append(deltas, models.PositionDelta{UserID: co.Order.UserID, Symbol: co.Order.Symbol, Qty: co.ClaimedQty})
		}
	}
	return posRepo.RevertReservations(ctx, deltas)
}

func revertOrdersAndCreateOutboxEvents(ctx context.Context, obr ports.OutboxRepository, event models.MatchingEvent, qty int, err error) error {
	event.Order.RemainingQuantity += qty
	updateStatus(&event.Order)
	event.Order.Status = "canceled"
	var orderEvents = []*models.OrderEvent{{
		Topic: "OrderEvents",
		Event: "OrderConfirmationFailed",
		TraceID: event.TraceID,
		Order: event.Order,
		Error: err.Error(),
	}}

	for _, o := range event.Orders {
		o.Order.RemainingQuantity += o.ClaimedQty
		updateStatus(&o.Order)
		returningEvent := createOrderEvent(event, &o.Order)
		orderEvents = append(orderEvents, &returningEvent)
	}

	return obr.CreateOrderEvents(ctx, orderEvents) // this commits the transaction for the failure path
}

func updateStatus(order *models.Order) {
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

func getQuantityNeeded(records []*models.ExecutionRecord) int {
	qty := 0

	for _, r := range records {
		qty += r.Quantity
	}

	return qty
}
