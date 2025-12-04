package handler_adapters

import (
	"brokerx/order-service/common"
	"brokerx/order-service/core"
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"context"
	"errors"
)

type OrderFailedHandler struct {
	OrderService core.OrderService
	Producer     ports.EventProducer
}

func (h *OrderFailedHandler) handle(ctx context.Context, event models.OrderEvent) error {
	log := common.FromContext(ctx)

	event.Order.Status = "canceled"
	err := h.OrderService.UpdateOrder(ctx, &event.Order)
	if err != nil {
		log.Error("Order could not be canceled", "error", err)
		return err
	} else {
		event.Event = "OrderCanceled"
		return h.Producer.SendEvent(ctx, "NotificationEvents", event.Event, event.Order, errors.New(event.Error))
	}
}
