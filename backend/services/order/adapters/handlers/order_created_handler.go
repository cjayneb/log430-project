package handler_adapters

import (
	"brokerx/order-service/common"
	"brokerx/order-service/core"
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"context"
)

type OrderCreatedHandler struct {
	ComplianceService core.ComplianceService
	Producer          ports.EventProducer
}

func (h *OrderCreatedHandler) handle(ctx context.Context, event models.OrderEvent) error {
	log := common.FromContext(ctx)

	err := h.ComplianceService.VerifyOrderCompliance(ctx, &event.Order)
	if err != nil {
		log.Error("Order not compliant", "error", err)
		event.Event = "OrderComplianceFailed"
		event.Order.Status = "canceled"
		return h.Producer.SendEvent(ctx, "OrderEvents", event.Event, event.Order, err)
	} else {
		event.Event = "OrderCompliant"
		log.Info("sending OrderCompliant Event")
		return h.Producer.SendEvent(ctx, "PortfolioEvents", event.Event, event.Order, nil)
	}
}
