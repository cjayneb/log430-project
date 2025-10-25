package ports

import (
	"brokerx/models"
)

type ExecutionQueue interface {
	EnqueueExecutionRecords(records []*models.ExecutionRecord) error
	DequeueExecutionRecords(batchSize int) ([]*models.ExecutionRecord, error)
}
