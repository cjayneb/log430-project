package ports

import (
	"brokerx/portfolio-service/models"
	"context"
	"time"
)

type OutboxRepository interface {
	CreateOrderEvents(ctx context.Context, events []*models.OrderEvent) error
	FetchPending(ctx context.Context, limit int) ([]models.OutboxRecord, error)
	MarkPublished(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64, errMsg string) error
	IncrementRetry(ctx context.Context, id int64, next time.Time) error
}
