package ports

import (
	"context"
)

type TransactionManager interface {
	Do(ctx context.Context, fn func(OrderRepository, ExecutionRepository) error) error
}
