package ports

import (
	"brokerx/order-service/models"
	"context"
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) (int, error)
	UpdateBatch(ctx context.Context, orders []*models.Order) error
	FindByUserId(ctx context.Context, userId int) ([]*models.Order, error)
}
