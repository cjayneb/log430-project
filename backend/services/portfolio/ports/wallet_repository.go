package ports

import (
	"brokerx/portfolio-service/models"
	"context"
)

type WalletRepository interface {
	FindByUserId(ctx context.Context, userId int) (*models.Wallet, error)
	AddFunds(ctx context.Context, userId int, amount float64) error
	ReserveFunds(ctx context.Context, userId int, amount float64) error
	RevertFundReservation(ctx context.Context, userId int, amount float64) error
	ReleaseFunds(ctx context.Context, deltas map[int]models.WalletDelta) (map[int]models.WalletDelta, error)
}
