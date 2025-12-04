package handler_adapters

import (
	"brokerx/notification-service/core"
	"brokerx/notification-service/models"
	"brokerx/notification-service/ports"
	"brokerx/notification-service/util"
	"context"
)

type OrderCompletedHandler struct {
	Service  core.NotificationService
	Producer ports.EventProducer
}

func (h *OrderCompletedHandler) handle(ctx context.Context, event models.OrderEvent) error {
	log := util.FromContext(ctx)

	if err := h.Service.SendNotification(ctx, event.Order); err != nil {
		log.Error("error when sending notification", "error", err, "orderId", event.Order.ID)
		event.Event = "OrderSagaCompleted"
		event.Topic = "OrderEvents"
		event.Error = err.Error()
		return h.Producer.SendEvent(ctx, event)
	}

	event.Event = "OrderSagaCompleted"
	event.Topic = "OrderEvents"
	return h.Producer.SendEvent(ctx, event)
}
