package ports

import "brokerx/models"

type MarketDataProvider interface {
	GetInstrumentBySymbol(symbol string) (*models.Instrument, error)
	GetCurrentStockPriceBySymbol(symbol string) (float64, error)
}