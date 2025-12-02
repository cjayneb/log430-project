package ports

import (
	"brokerx/portfolio-service/models"
	"context"
)

type PositionRepository interface {
	FindByUserIdAndSymbol(ctx context.Context, userId int, symbol string) ([]*models.Position, error)
	FindByUserId(ctx context.Context, userId int) ([]*models.Position, error)
	ReleaseQuantity(ctx context.Context, deltas []*models.ClaimedCandidate) error
	ReserveQuantity(ctx context.Context, deltas []models.PositionDelta) error
	RevertReservations(ctx context.Context, deltas []models.PositionDelta) error
	AddAvailableQuantity(ctx context.Context, deltas []*models.ClaimedCandidate) error
}
