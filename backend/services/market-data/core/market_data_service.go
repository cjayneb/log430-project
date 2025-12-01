package core

import (
	"brokerx/market-data-service/models"
	"brokerx/market-data-service/util"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

const INSTRUMENT_SOURCE_FILE string = "instruments.json"
const PRICES_SOURCE_FILE string = "prices.json"

type MarketDataService interface {
	GetInstrumentBySymbol(ctx context.Context, symbol string) (*models.Instrument, error)
	GetCurrentStockPriceBySymbol(ctx context.Context, symbol string) (float64, error)
}

type MarketDataServiceImpl struct {
	Instruments []models.Instrument
	Prices      []Price
}

type Price struct {
	Symbol string  `json:"Symbol"`
	Price  float64 `json:"Price"`
}

func NewMarketDataServiceImpl(resourcesPath string) *MarketDataServiceImpl {
	if resourcesPath == "" {
		resourcesPath = "resources/"
	}
	fileContent, err := os.ReadFile(resourcesPath + INSTRUMENT_SOURCE_FILE)
	if err != nil {
		slog.Error("error reading file", "error", err)
		os.Exit(1)
	}
	var instruments []models.Instrument
	err = json.Unmarshal(fileContent, &instruments)
	if err != nil {
		slog.Error("error unmarshaling JSON", "error", err)
		os.Exit(1)
	}

	fileContent, err = os.ReadFile(resourcesPath + PRICES_SOURCE_FILE)
	if err != nil {
		slog.Error("error reading file", "error", err)
		os.Exit(1)
	}
	var prices []Price
	err = json.Unmarshal(fileContent, &prices)
	if err != nil {
		slog.Error("error unmarshaling JSON", "error", err)
		os.Exit(1)
	}

	return &MarketDataServiceImpl{Instruments: instruments, Prices: prices}
}

func (m *MarketDataServiceImpl) GetCurrentStockPriceBySymbol(ctx context.Context, symbol string) (float64, error) {
	log := util.FromContext(ctx)

	for _, p := range m.Prices {
		if p.Symbol == symbol {
			return p.Price, nil
		}
	}

	log.Warn(fmt.Sprintf("price for symbol {%s} not found", symbol))
	return 0.0, fmt.Errorf("price for symbol {%s} not found", symbol)
}

func (m *MarketDataServiceImpl) GetInstrumentBySymbol(ctx context.Context, symbol string) (*models.Instrument, error) {
	log := util.FromContext(ctx)

	for _, i := range m.Instruments {
		if i.Symbol == symbol {
			return &i, nil
		}
	}

	log.Warn(fmt.Sprintf("instrument for symbol {%s} not found", symbol))
	return nil, fmt.Errorf("instrument for symbol {%s} not found", symbol)
}

var _ MarketDataService = (*MarketDataServiceImpl)(nil) // Ensure interface is implemented at compile time
