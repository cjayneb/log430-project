package ports

import (
	"brokerx/portfolio-service/models"
	"context"
)

type ExecutionRepository interface {
	CreateBatch(ctx context.Context, execs []*models.ExecutionRecord) error
}
