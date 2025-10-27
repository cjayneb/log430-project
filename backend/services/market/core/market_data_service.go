package core

import (
	"brokerx/market-data-service/models"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

const INSTRUMENT_SOURCE_FILE string = "instruments.json"
const PRICES_SOURCE_FILE string = "prices.json"

type MarketDataService interface {
	GetInstrumentBySymbol(symbol string) (*models.Instrument, error)
	GetCurrentStockPriceBySymbol(symbol string) (float64, error)
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
		resourcesPath = "../adapters/resources/"
	}
	fileContent, err := os.ReadFile(resourcesPath + INSTRUMENT_SOURCE_FILE)
	if err != nil {
		log.Fatalf("error reading file : %v", err)
	}
	var instruments []models.Instrument
	err = json.Unmarshal(fileContent, &instruments)
	if err != nil {
		log.Fatalf("error unmarshaling JSON : %v", err)
	}

	fileContent, err = os.ReadFile(resourcesPath + PRICES_SOURCE_FILE)
	if err != nil {
		log.Fatalf("error reading file : %v", err)
	}
	var prices []Price
	err = json.Unmarshal(fileContent, &prices)
	if err != nil {
		log.Fatalf("error unmarshaling JSON : %v", err)
	}

	return &MarketDataServiceImpl{Instruments: instruments, Prices: prices}
}

func (m *MarketDataServiceImpl) GetCurrentStockPriceBySymbol(symbol string) (float64, error) {
	for _, p := range m.Prices {
		if strings.EqualFold(p.Symbol, symbol) {
			return p.Price, nil
		}
	}

	return 0.0, fmt.Errorf("price for symbol {%s} not found", symbol)
}

func (m *MarketDataServiceImpl) GetInstrumentBySymbol(symbol string) (*models.Instrument, error) {
	for _, i := range m.Instruments {
		if strings.EqualFold(i.Symbol, symbol) {
			return &i, nil
		}
	}

	return nil, fmt.Errorf("instrument for symbol {%s} not found", symbol)
}

var _ MarketDataService = (*MarketDataServiceImpl)(nil) // Ensure interface is implemented at compile time
