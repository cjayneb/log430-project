package handler_adapters

import (
	"brokerx/matching-service/models"
	"brokerx/matching-service/ports"
	"context"
)

type OrderOpenHandler struct {
	OrderBook ports.OrderBook
}

func (h *OrderOpenHandler) handle(ctx context.Context, event models.OrderEvent) {
	h.OrderBook.Return(ctx, []*models.Order{&event.Order})
}
