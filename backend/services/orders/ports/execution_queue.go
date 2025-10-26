package ports

import (
	"brokerx/order-service/models"
)

type ExecutionQueue interface {
	EnqueueExecutionRecords(records []*models.ExecutionRecord) error
	DequeueExecutionRecords(batchSize int) ([]*models.ExecutionRecord, error)
}
