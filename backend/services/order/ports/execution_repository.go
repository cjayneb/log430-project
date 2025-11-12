package ports

import (
	"brokerx/order-service/models"
)

type ExecutionRepository interface {
	CreateBatch(execs []*models.ExecutionRecord) error
}
