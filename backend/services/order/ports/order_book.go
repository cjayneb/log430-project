package ports

import (
	"brokerx/order-service/models"
	"context"
)

type OrderBook interface {
	FetchByIDs(ctx context.Context, ids []string) ([]*models.Order, error)
	DequeueOrders(ctx context.Context, batchSize int) ([]*models.Order, error)
}
