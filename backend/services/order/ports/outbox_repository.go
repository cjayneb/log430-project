package ports

import (
	"brokerx/order-service/models"
	"context"
)

type OutboxRepository interface {
	CreateOrderEvents(ctx context.Context, events []*models.OrderEvent) error
}
