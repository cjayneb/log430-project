package handler_adapters

import (
	"brokerx/portfolio-service/models"
	"brokerx/portfolio-service/ports"
	"brokerx/portfolio-service/util"
	"context"
)

type OrderCompliantHandler struct {
	Producer ports.EventProducer
	Tm       ports.TransactionManager
}

func (h *OrderCompliantHandler) handle(ctx context.Context, event models.OrderEvent) error {
	log := util.FromContext(ctx)

	order := event.Order
	err := h.Tm.Do(ctx, func(_ ports.ExecutionRepository, wr ports.WalletRepository, pr ports.PositionRepository, obr ports.OutboxRepository) error {
		total := order.UnitPrice * float64(order.Quantity)
		err := reserveFunds(ctx, order, wr, total)
		if err != nil {
			log.Error("error reserving funds", "error", err)
			return createReservingQuantitiesFailedOutboxEvent(ctx, event, obr, err)
		}

		err = reservePosition(ctx, order, pr)
		if err != nil {
			log.Error("error reserving position quantity for order", "error", err)
			if err := releaseFunds(ctx, order, wr, total); err != nil {
				log.Error("error reverting fund reservation", "error", err)
				return err
			}
			return createReservingQuantitiesFailedOutboxEvent(ctx, event, obr, err)
		}

		if err := createQuantitiesReservedOutboxEvent(ctx, event, obr); err != nil {
			log.Error("error creating QuantitiesReserved outbox event", "error", err)
			return err
		}

		return nil // this commits the transaction for the successful path
	})

	if err != nil {
		log.Error("error processing event. OrderCompliant event will be consumed until successful...", "orderId", event.Order.ID, "error", err)
		return err
	}

	// Dont send any event. Outbox dispatcher will send events asynchronously.
	return nil
}

func reserveFunds(ctx context.Context, order models.Order, walletRepo ports.WalletRepository, total float64) error {
	if order.Action == "sell" || order.Type == "market" {
		return nil
	}

	return walletRepo.ReserveFunds(ctx, order.UserID, total)
}

func releaseFunds(ctx context.Context, order models.Order, walletRepo ports.WalletRepository, total float64) error {
	if order.Action == "sell" || order.Type == "market" {
		return nil
	}

	return walletRepo.RevertFundReservation(ctx, order.UserID, total)
}

func reservePosition(ctx context.Context, order models.Order, posRepo ports.PositionRepository) error {
	if order.Action == "buy" {
		return nil
	}

	posDelta := []models.PositionDelta{{UserID: order.UserID, Symbol: order.Symbol, Qty: order.Quantity}}
	return posRepo.ReserveQuantity(ctx, posDelta)
}

func createQuantitiesReservedOutboxEvent(ctx context.Context, event models.OrderEvent, obr ports.OutboxRepository) error {
	event.Topic = "MatchingEvents"
	event.Event = "QuantitiesReserved"
	return obr.CreateOrderEvents(ctx, []*models.OrderEvent{&event})
}

func createReservingQuantitiesFailedOutboxEvent(ctx context.Context, event models.OrderEvent, obr ports.OutboxRepository, err error) error {
	event.Topic = "OrderEvents"
	event.Event = "ReservingQuantitiesFailed"
	event.Order.Status = "canceled"
	event.Error = err.Error()
	return obr.CreateOrderEvents(ctx, []*models.OrderEvent{&event})
}
