package adapters

import (
	"brokerx/models"
	"brokerx/ports"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const INSTRUMENT_SOURCE_FILE_PATH string = "../../ressources/instruments.json"
const PRICES_SOURCE_FILE_PATH string = "../../ressources/prices.json"

type Price struct {
	Symbol string `json:"Symbol"`
	Price float64 `json:"Price"`
}

type MarketDataProvider struct {
	Instruments []models.Instrument
	Prices []Price
}

func NewMarketDataProvider() *MarketDataProvider {
	fileContent, err := os.ReadFile(INSTRUMENT_SOURCE_FILE_PATH)
	if err != nil {
		log.Fatalf("error reading file (%s): %v", INSTRUMENT_SOURCE_FILE_PATH, err)
	}
	var instruments []models.Instrument
	err = json.Unmarshal(fileContent, &instruments)
	if err != nil {
		log.Fatalf("error unmarshaling JSON (%s): %v", INSTRUMENT_SOURCE_FILE_PATH, err)
	}

	fileContent, err = os.ReadFile(PRICES_SOURCE_FILE_PATH)
	if err != nil {
		log.Fatalf("error reading file (%s): %v", PRICES_SOURCE_FILE_PATH, err)
	}
	var prices []Price
	err = json.Unmarshal(fileContent, &prices)
	if err != nil {
		log.Fatalf("error unmarshaling JSON (%s): %v", PRICES_SOURCE_FILE_PATH, err)
	}

	return &MarketDataProvider{Instruments: instruments, Prices: prices}
}

func (provider * MarketDataProvider) GetInstrumentBySymbol(symbol string) (*models.Instrument, error) {
	for _, i := range provider.Instruments {
		if i.Symbol == symbol {
			return &i, nil
		}
	}

	return nil, fmt.Errorf("error : instrument for symbol {%s} not found", symbol)
}

func (provider * MarketDataProvider) GetCurrentStockPriceBySymbol(symbol string) (float64, error) {
	for _, p := range provider.Prices {
		if p.Symbol == symbol {
			return p.Price, nil
		}
	}

	return 0.0, fmt.Errorf("error : price for symbol {%s} not found", symbol)
}

var _ ports.MarketDataProvider = (*MarketDataProvider)(nil) // Ensure interface is implemented at compile time