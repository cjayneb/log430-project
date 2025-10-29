package mocks

import "brokerx/market-data-service/models"

type MockMarketDataService struct {
	Instrument *models.Instrument
	Price      float64
	Err        error
}

func (m *MockMarketDataService) GetCurrentStockPriceBySymbol(symbol string) (float64, error) {
	if m.Err != nil {
		return 0.0, m.Err
	}
	return m.Price, nil
}

func (m *MockMarketDataService) GetInstrumentBySymbol(symbol string) (*models.Instrument, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Instrument, nil
}
