package core

import (
	"brokerx/portfolio-service/models"
	"brokerx/portfolio-service/ports"
	"fmt"
)

type PortfolioService interface {
	GetWallet(userId int) (*models.Wallet, error)
	FetchPositions(userId int, symbol string) ([]*models.Position, error)
}

type PortfolioServiceImpl struct {
	PositionsRepo ports.PositionRepository
	WalletRepo    ports.WalletRepository
}

func (service *PortfolioServiceImpl) GetWallet(userId int) (*models.Wallet, error) {
	wallet, err := service.WalletRepo.FindByUserId(userId)
	if wallet == nil {
		return wallet, fmt.Errorf("no wallet found for user %d", userId)
	}
	return wallet, err
}

func (service *PortfolioServiceImpl) FetchPositions(userId int, symbol string) ([]*models.Position, error) {
	return service.PositionsRepo.FindByUserIdAndSymbol(userId, symbol)
}

var _ PortfolioService = (*PortfolioServiceImpl)(nil) // Ensure interface is implemented at compile time
