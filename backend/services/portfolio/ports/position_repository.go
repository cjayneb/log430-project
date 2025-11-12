package ports

import (
	"brokerx/portfolio-service/models"
	"context"
)

type PositionRepository interface {
	FindByUserIdAndSymbol(ctx context.Context, userId int, symbol string) ([]*models.Position, error)
}
