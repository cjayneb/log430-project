package ports

import (
	"brokerx/matching-service/models"
	"context"
)

type ExecutionQueue interface {
	EnqueueExecutionRecords(ctx context.Context, records []*models.ExecutionRecord) error
}
