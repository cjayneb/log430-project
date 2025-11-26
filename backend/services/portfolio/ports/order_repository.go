package ports

import (
	"brokerx/portfolio-service/models"
	"context"
)

type OrderRepository interface {
	UpdateBatch(ctx context.Context, orders []*models.Order) error
}
