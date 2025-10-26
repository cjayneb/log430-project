package ports

import (
	"brokerx/order-service/models"
	"context"
)

type TransactionManager interface {
	Do(ctx context.Context, fn func(OrderRepository, ExecutionRepository) error) error
	DoReadOnly(ctx context.Context, fn func(OrderRepository) ([]*models.Order, error)) ([]*models.Order, error)
}
