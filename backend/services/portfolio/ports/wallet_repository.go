package ports

import (
	"brokerx/portfolio-service/models"
	"context"
)

type WalletRepository interface {
	FindByUserId(ctx context.Context, userId int) (*models.Wallet, error)
	AddFunds(ctx context.Context, userId int, amount float64) error
}
