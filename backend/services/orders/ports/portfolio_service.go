package ports

import (
	"brokerx/order-service/models"
	"context"
)

type PortfolioService interface {
	FetchPositions(ctx context.Context, userId, symbol string) ([]*models.Position, error)
	GetWallet(ctx context.Context, userId string) (*models.Wallet, error)
}
