package adapters

import (
	"brokerx/models"
	"brokerx/ports"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const INSTRUMENT_SOURCE_FILE string = "instruments.json"
const PRICES_SOURCE_FILE string = "prices.json"

type Price struct {
	Symbol string `json:"Symbol"`
	Price float64 `json:"Price"`
}

type MarketDataProvider struct {
	Instruments []models.Instrument
	Prices []Price
}

func NewMarketDataProvider(resourcesPath string) *MarketDataProvider {
	if resourcesPath == "" {
		resourcesPath = "../adapters/resources/"
	}
	fileContent, err := os.ReadFile(resourcesPath+INSTRUMENT_SOURCE_FILE)
	if err != nil {
		log.Fatalf("error reading file : %v", err)
	}
	var instruments []models.Instrument
	err = json.Unmarshal(fileContent, &instruments)
	if err != nil {
		log.Fatalf("error unmarshaling JSON : %v", err)
	}

	fileContent, err = os.ReadFile(resourcesPath+PRICES_SOURCE_FILE)
	if err != nil {
		log.Fatalf("error reading file : %v", err)
	}
	var prices []Price
	err = json.Unmarshal(fileContent, &prices)
	if err != nil {
		log.Fatalf("error unmarshaling JSON : %v", err)
	}

	return &MarketDataProvider{Instruments: instruments, Prices: prices}
}

func (provider * MarketDataProvider) GetInstrumentBySymbol(symbol string) (*models.Instrument, error) {
	for _, i := range provider.Instruments {
		if i.Symbol == symbol {
			return &i, nil
		}
	}

	return nil, fmt.Errorf("instrument for symbol {%s} not found", symbol)
}

func (provider * MarketDataProvider) GetCurrentStockPriceBySymbol(symbol string) (float64, error) {
	for _, p := range provider.Prices {
		if p.Symbol == symbol {
			return p.Price, nil
		}
	}

	return 0.0, fmt.Errorf("price for symbol {%s} not found", symbol)
}

var _ ports.MarketDataProvider = (*MarketDataProvider)(nil) // Ensure interface is implemented at compile time