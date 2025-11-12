package ports

import (
	"brokerx/order-service/models"
	"context"
)

type ExecutionQueue interface {
	DequeueExecutionRecords(ctx context.Context, batchSize int) ([]*models.ExecutionRecord, error)
}
