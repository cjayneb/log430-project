package handler_adapters

import (
	"brokerx/order-service/common"
	"brokerx/order-service/core"
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"context"
)

type UpdateOrderHandler struct {
	OrderService core.OrderService
	Tm           ports.TransactionManager
}

func (h *UpdateOrderHandler) handle(ctx context.Context, event models.OrderEvent) error {
	log := common.FromContext(ctx)

	err := h.Tm.Do(ctx, func(or ports.OrderRepository, obr ports.OutboxRepository) error {
		if err := or.UpdateBatch(ctx, []*models.Order{&event.Order}); err != nil {
			log.Error("error when updating order", "error", err)
			return err
		}

		return obr.CreateOrderEvents(ctx, []*models.OrderEvent{createOrderEvent(&event)})
	})

	if err != nil {
		log.Error("error processing update order event. Will retry until successful", "error", err)
		return err
	}

	log.Info("successfully processed update order event", "orderId", event.Order.ID, "event", event.Event)
	return nil
}

func createOrderEvent(event *models.OrderEvent) *models.OrderEvent {
	topic := "MatchingEvents"
	eventType := "OrderOpen"
	if event.Order.Status == "filled" || event.Order.Status == "canceled" || (event.Order.Timing == "ioc" && event.Order.Status == "partially_filled") {
		eventType = "OrderCompleted"
		topic = "NotificationEvents"
	}
	event.Topic = topic
	event.Event = eventType
	return event
}
