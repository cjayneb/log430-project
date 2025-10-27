package ports

import (
	"brokerx/order-service/models"
	"context"
)

type MarketDataProvider interface {
	GetInstrumentBySymbol(ctx context.Context, symbol string) (*models.Instrument, error)
	GetCurrentStockPriceBySymbol(ctx context.Context, symbol string) (float64, error)
}
