package ports

import (
	"brokerx/order-service/models"
	"context"
)

type MarketDataProvider interface {
	GetInstrumentBySymbol(ctx context.Context, symbol string) (*models.Instrument, error)
	GetCurrentStockPriceBySymbol(ctx context.Context, symbol string) (float64, error)
}

type MarketDataProviderImpl struct{}

func (mdp *MarketDataProviderImpl) GetInstrumentBySymbol(ctx context.Context, symbol string) (*models.Instrument, error) {
	return nil, nil
}

func (mdp *MarketDataProviderImpl) GetCurrentStockPriceBySymbol(ctx context.Context, symbol string) (float64, error) {
	return 0, nil
}

var _ MarketDataProvider = (*MarketDataProviderImpl)(nil) // Ensure interface is implemented at compile time
