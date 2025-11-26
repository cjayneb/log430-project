package handler_adapters

import (
	"brokerx/matching-service/core"
	"brokerx/matching-service/models"
	"brokerx/matching-service/ports"
	"brokerx/matching-service/util"
	"context"
)

type OrderValidatedHandler struct {
	MatchingService core.MatchingEngine
	Producer        ports.EventProducer
}

func (h *OrderValidatedHandler) handle(ctx context.Context, event models.OrderEvent) error {
	log := util.FromContext(ctx)

	err := h.MatchingService.QueueOrder(ctx, &event.Order)
	if err != nil {
		log.Error("could not queue order", "error", err)
		return h.Producer.SendEvent(ctx, "OrderEvents", "OrderMatchingFailed", event.Order, err)
	}

	err = h.MatchingService.SubmitOrder(ctx, event.Order.ID)
	if err != nil {
		log.Error("error submitting order to matching engine", "error", err)
		return h.Producer.SendEvent(ctx, "OrderEvents", "OrderMatchingFailed", event.Order, err)
	}

	return nil
}
