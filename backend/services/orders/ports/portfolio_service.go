package ports

import (
	"brokerx/order-service/models"
	"fmt"
)

type PortfolioService interface {
	FetchPositions(userId, symbol string) ([]*models.Position, error)
	GetWallet(userId string) (*models.Wallet, error)
}

type PortfolioServiceImpl struct {
}

func (p *PortfolioServiceImpl) FetchPositions(userId, symbol string) ([]*models.Position, error) {
	return nil, fmt.Errorf("not implemented")
}

func (p *PortfolioServiceImpl) GetWallet(userId string) (*models.Wallet, error) {
	return nil, fmt.Errorf("not implemented")
}

var _ PortfolioService = (*PortfolioServiceImpl)(nil) // Ensure interface is implemented at compile time
