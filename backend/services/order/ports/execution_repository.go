package ports

import "brokerx/order-service/models"

type ExecutionRepository interface {
	Create(record *models.ExecutionRecord) error
	CreateBatch(execs []*models.ExecutionRecord) error
}
