package client_adapters

import (
	"brokerx/order-service/common"
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
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
	return &MarketDataProviderImpl{
		BaseUrl: baseUrl + MARKET_DATA_SERVICE_BASE_PATH,
		Client:  common.SharedHttpClient,
	}
}

func (m *MarketDataProviderImpl) GetCurrentStockPriceBySymbol(ctx context.Context, symbol string) (float64, error) {
	log := common.FromContext(ctx)

	req, err := common.MakeAuthenticatedRequest(ctx, "GET", m.BaseUrl+STOCK_PRICE_ENDPOINT+"?symbol="+symbol, nil)
	if err != nil {
		msg := "error creating request to fetch stock price from market data service"
		log.Error(msg, "error", err)
		return 0.0, errors.New(msg)
	}

	resp, err := m.Client.Do(req)
	if err != nil {
		msg := "error making request to market data service"
		log.Error(msg, "error", err)
		return 0.0, errors.New(msg)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg := "market data service returned non-200 status"
		log.Error(msg, "status", resp.Status)
		return 0.0, errors.New(msg)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		msg := "error reading response body"
		log.Error(msg, "error", err)
		return 0.0, errors.New(msg)
	}

	var priceResp StockPriceResponse
	err = json.Unmarshal(body, &priceResp)
	if err != nil {
		msg := "error unmarshaling JSON"
		log.Error(msg, "error", err)
		return 0.0, errors.New(msg)
	}

	return priceResp.Price, nil
}

func (m *MarketDataProviderImpl) GetInstrumentBySymbol(ctx context.Context, symbol string) (*models.Instrument, error) {
	log := common.FromContext(ctx)

	req, err := common.MakeAuthenticatedRequest(ctx, "GET", m.BaseUrl+STOCK_INSTRUMENT_ENDPOINT+"?symbol="+symbol, nil)
	if err != nil {
		msg := "error creating request to fetch instrument from market data service"
		log.Error(msg, "error", err)
		return nil, errors.New(msg)
	}

	resp, err := m.Client.Do(req)
	if err != nil {
		msg := "error making request to market data service"
		log.Error(msg, "error", err)
		return nil, errors.New(msg)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg := "market data service returned non-200 status when fetching instrument"
		log.Error(msg, "status", resp.Status)
		return nil, errors.New(msg)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		msg := "error reading response body"
		log.Error(msg, "error", err)
		return nil, errors.New(msg)
	}

	var instrumentResp InstrumentResponse
	err = json.Unmarshal(body, &instrumentResp)
	if err != nil {
		msg := "error unmarshaling JSON"
		log.Error(msg, "error", err)
		return nil, errors.New(msg)
	}

	return &instrumentResp.Instrument, nil
}

var _ ports.MarketDataProvider = (*MarketDataProviderImpl)(nil) // Ensure interface is implemented at compile time
