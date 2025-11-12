package ports

import (
	"brokerx/matching-service/models"
	"context"
)

type OrderBook interface {
	GetById(ctx context.Context, orderId int) (models.Order, error)
	FindMatchesLimit(ctx context.Context, symbol string, action string, unitPrice float64, batchSize int) ([]*models.Order, error)
	FindMatchesMarket(ctx context.Context, symbol string, action string, batchSize int) ([]*models.Order, error)
	Insert(ctx context.Context, order *models.Order) error
	Return(ctx context.Context, orders []*models.Order) error
	EnqueueOrders(ctx context.Context, orders []*models.Order) error
}
