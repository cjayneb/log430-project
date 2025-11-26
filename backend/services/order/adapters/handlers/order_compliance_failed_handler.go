package handler_adapters

import (
	"brokerx/order-service/common"
	"brokerx/order-service/core"
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"context"
)

type OrderComplianceFailedHandler struct {
	OrderService core.OrderService
	Producer     ports.EventProducer
}

func (h *OrderComplianceFailedHandler) handle(ctx context.Context, event models.OrderEvent) error {
	log := common.FromContext(ctx)

	event.Order.Status = "canceled"
	err := h.OrderService.UpdateOrder(ctx, &event.Order)
	if err != nil {
		log.Error("Order could not be canceled", "error", err)
		event.Event = "OrderSagaCompleted"
		return h.Producer.SendEvent(ctx, "OrderEvents", event.Event, event.Order, err)
	} else {
		event.Event = "OrderSagaCompleted"
		return h.Producer.SendEvent(ctx, "OrderEvents", event.Event, event.Order, nil)
	}
}
