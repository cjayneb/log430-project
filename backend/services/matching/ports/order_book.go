package ports

import (
	"brokerx/matching-service/models"
	"context"
)

type OrderBook interface {
	FindMatches(ctx context.Context, order *models.Order, batchSize int) ([]*models.Order, error)
	Insert(ctx context.Context, order *models.Order) error
	Return(ctx context.Context, orders []*models.Order)
}
