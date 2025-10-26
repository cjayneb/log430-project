package ports

import "brokerx/order-service/models"

type MarketDataProvider interface {
	GetInstrumentBySymbol(symbol string) (*models.Instrument, error)
	GetCurrentStockPriceBySymbol(symbol string) (float64, error)
}

type MarketDataProviderImpl struct{}

func (mdp *MarketDataProviderImpl) GetInstrumentBySymbol(symbol string) (*models.Instrument, error) {
	return nil, nil
}

func (mdp *MarketDataProviderImpl) GetCurrentStockPriceBySymbol(symbol string) (float64, error) {
	return 0, nil
}

var _ MarketDataProvider = (*MarketDataProviderImpl)(nil) // Ensure interface is implemented at compile time
