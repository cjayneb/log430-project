package client_adapters

import (
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const MARKET_DATA_SERVICE_BASE_PATH = "/api/market"
const STOCK_PRICE_ENDPOINT = "/stock/price"
const STOCK_INSTRUMENT_ENDPOINT = "/stock/instrument"

type MarketDataProviderImpl struct {
	BaseUrl string
	Client  *http.Client
}

type StockPriceResponse struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
}

type InstrumentResponse struct {
	Instrument models.Instrument `json:"instrument"`
}

func NewMarketDataProvider(baseUrl string) *MarketDataProviderImpl {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	return &MarketDataProviderImpl{
		BaseUrl: baseUrl + MARKET_DATA_SERVICE_BASE_PATH,
		Client:  client,
	}
}

func (m *MarketDataProviderImpl) GetCurrentStockPriceBySymbol(ctx context.Context, symbol string) (float64, error) {
	req, err := makeAuthenticatedRequest(ctx, "GET", m.BaseUrl+STOCK_PRICE_ENDPOINT+"?symbol="+symbol)
	if err != nil {
		return 0.0, err
	}

	resp, err := m.Client.Do(req)
	if err != nil {
		return 0.0, fmt.Errorf("error making request to market data service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0.0, fmt.Errorf("market data service returned non-200 status: %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0.0, fmt.Errorf("error reading response body: %v", err)
	}

	var priceResp StockPriceResponse
	err = json.Unmarshal(body, &priceResp)
	if err != nil {
		return 0.0, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	return priceResp.Price, nil
}

func (m *MarketDataProviderImpl) GetInstrumentBySymbol(ctx context.Context, symbol string) (*models.Instrument, error) {
	req, err := makeAuthenticatedRequest(ctx, "GET", m.BaseUrl+STOCK_INSTRUMENT_ENDPOINT+"?symbol="+symbol)
	if err != nil {
		return nil, err
	}

	resp, err := m.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to market data service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("market data service returned non-200 status: %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var instrumentResp InstrumentResponse
	err = json.Unmarshal(body, &instrumentResp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	return &instrumentResp.Instrument, nil
}

var _ ports.MarketDataProvider = (*MarketDataProviderImpl)(nil) // Ensure interface is implemented at compile time
