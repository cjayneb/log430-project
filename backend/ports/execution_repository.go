package ports

import "brokerx/models"

type ExecutionRepository interface {
	Create(record *models.ExecutionRecord) error
	CreateBatch(execs []*models.ExecutionRecord) error
}
