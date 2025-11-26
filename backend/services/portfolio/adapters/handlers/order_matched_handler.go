package handler_adapters

import (
	"brokerx/portfolio-service/models"
	"brokerx/portfolio-service/ports"
	"brokerx/portfolio-service/util"
	"context"
	"errors"
)

type OrderMatchedHandler struct {
	ExecutionRepo ports.ExecutionRepository
	OrderRepo     ports.OrderRepository
	PositionRepo  ports.PositionRepository
	WalletRepo    ports.WalletRepository
	Producer      ports.EventProducer
}

func (h *OrderMatchedHandler) handle(ctx context.Context, event models.MatchingEvent) error {
	log := util.FromContext(ctx)

	if len(event.Executions) == 0 {
		return h.successEvent(ctx, &event.Order)
	}

	if err := h.validateFunds(ctx, &event); err != nil {
		log.Error("funds validation failed", "error", err)
		return h.failEvent(ctx, &event, err)
	}

	if err := h.ExecutionRepo.CreateBatch(ctx, event.Executions); err != nil {
		log.Error("execution batch failed", "error", err)
		return h.failEvent(ctx, &event, err)
	}

	// TODO: Position update logic here
	// if err := h.PositionRepo.Update(...); err != nil { ... }

	event.Orders = append(event.Orders, &event.Order)
	if err := h.OrderRepo.UpdateBatch(ctx, event.Orders); err != nil {
		log.Error("order batch update failed", "error", err)
		return h.failEvent(ctx, &event, err)
	}

	for _, o := range event.Orders {
		if err := h.successEvent(ctx, o); err != nil {
			log.Error("error sending order back", "error", err)
		}
	}

	return h.successEvent(ctx, &event.Order)
}

func (h *OrderMatchedHandler) successEvent(ctx context.Context, order *models.Order) error {
	topic := "MatchingEvents"
	event := "OrderOpen"
	if order.Status == "filled" {
		event = "SagaCompleted"
		topic = "OrderEvents"
	}
	return h.Producer.SendEvent(ctx, topic, event, *order, nil)
}

func (h *OrderMatchedHandler) failEvent(ctx context.Context, event *models.MatchingEvent, cause error) error {
	event.Event = "OrderConfirmationFailed"
	return h.Producer.SendEvent(ctx, "OrderEvents", event.Event, event.Order, cause)
}

func (h *OrderMatchedHandler) validateFunds(ctx context.Context, event *models.MatchingEvent) error {
	if event.Order.Action == "sell" {
		return nil
	}

	total, qty := getFundsNeeded(event.Executions)

	wallet, err := h.WalletRepo.FindByUserId(ctx, event.Order.UserID)
	if err != nil {
		event.Order.RemainingQuantity += qty
		return err
	}

	if wallet != nil && wallet.AvailableFunds-total < 0 { //TODO:change comparison to reservedFunds
		event.Order.RemainingQuantity += qty
		return errors.New("not enough funds")
	}

	return nil
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
