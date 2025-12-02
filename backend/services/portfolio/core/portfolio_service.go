package core

import (
	"brokerx/portfolio-service/models"
	"brokerx/portfolio-service/ports"
	"brokerx/portfolio-service/util"
	"context"
	"errors"
	"fmt"
)

type PortfolioService interface {
	GetWallet(ctx context.Context, userId int) (*models.Wallet, error)
	FundWallet(ctx context.Context, userId int, amount float64) error
	FetchPositionsForUser(ctx context.Context, userId int) ([]*models.Position, error)
	FetchPositionsForSymbol(ctx context.Context, userId int, symbol string) ([]*models.Position, error)
}

type PortfolioServiceImpl struct {
	PositionsRepo ports.PositionRepository
	WalletRepo    ports.WalletRepository
}

func (service *PortfolioServiceImpl) FundWallet(ctx context.Context, userId int, amount float64) error {
	log := util.FromContext(ctx)

	// TODO: add checks for compliance
	if amount <=0 {
		msg := "amount must be positive"
		log.Error(msg)
		return errors.New(msg)
	}
	return service.WalletRepo.AddFunds(ctx, userId, amount)
}

// TODO: Reserve funds and add to wallet ledger with idempotency key and transaction manager
func (service *PortfolioServiceImpl) GetWallet(ctx context.Context, userId int) (*models.Wallet, error) {
	wallet, err := service.WalletRepo.FindByUserId(ctx, userId)
	if wallet == nil {
		return wallet, fmt.Errorf("no wallet found for user %d", userId)
	}
	return wallet, err
}

func (service *PortfolioServiceImpl) FetchPositionsForSymbol(ctx context.Context, userId int, symbol string) ([]*models.Position, error) {
	return service.PositionsRepo.FindByUserIdAndSymbol(ctx, userId, symbol)
}

func (service *PortfolioServiceImpl) FetchPositionsForUser(ctx context.Context, userId int) ([]*models.Position, error) {
	return service.PositionsRepo.FindByUserId(ctx, userId)
}

var _ PortfolioService = (*PortfolioServiceImpl)(nil) // Ensure interface is implemented at compile time
