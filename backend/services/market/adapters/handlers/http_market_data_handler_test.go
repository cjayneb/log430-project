package handler_adapters_test

import (
	handler_adapters "brokerx/market-data-service/adapters/handlers"
	"brokerx/market-data-service/core"
	"brokerx/market-data-service/mocks"
	"brokerx/market-data-service/models"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMarketDataHandler_GetStockPrice(t *testing.T) {
	tests := []struct {
		name           string
		headers        map[string]string
		query          string
		mockService    core.MarketDataService
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "missing user header",
			headers:        map[string]string{},
			query:          "?symbol=AAPL",
			mockService:    &mocks.MockMarketDataService{Price: 100},
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       "missing user authentication context",
		},
		{
			name: "missing bearer token",
			headers: map[string]string{
				handler_adapters.USER_ID_HEADER_KEY: "user1",
				"Authorization":                     "abc",
			},
			query:          "?symbol=AAPL",
			mockService:    &mocks.MockMarketDataService{Price: 100},
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       "missing authorization token",
		},
		{
			name: "missing symbol query",
			headers: map[string]string{
				handler_adapters.USER_ID_HEADER_KEY: "user1",
				"Authorization":                     "Bearer token",
			},
			query:          "",
			mockService:    &mocks.MockMarketDataService{Price: 100},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "missing 'symbol' query parameter",
		},
		{
			name: "service returns error",
			headers: map[string]string{
				handler_adapters.USER_ID_HEADER_KEY: "user1",
				"Authorization":                     "Bearer token",
			},
			query:          "?symbol=AAPL",
			mockService:    &mocks.MockMarketDataService{Err: errors.New("db error")},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "failed to get price",
		},
		{
			name: "success",
			headers: map[string]string{
				handler_adapters.USER_ID_HEADER_KEY: "user1",
				"Authorization":                     "Bearer token",
			},
			query:          "?symbol=AAPL",
			mockService:    &mocks.MockMarketDataService{Price: 150.5},
			wantStatusCode: http.StatusOK,
			wantBody:       `"symbol":"AAPL"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := handler_adapters.MarketDataHandler{
				Service: tt.mockService,
			}

			req := httptest.NewRequest(http.MethodGet, "/price"+tt.query, nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			rr := httptest.NewRecorder()

			handler.GetStockPrice(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Errorf("expected status %d, got %d", tt.wantStatusCode, rr.Code)
			}
			if !strings.Contains(rr.Body.String(), tt.wantBody) {
				t.Errorf("expected body to contain %q, got %s", tt.wantBody, rr.Body.String())
			}
		})
	}
}

func TestMarketDataHandler_GetInstrument(t *testing.T) {
	tests := []struct {
		name           string
		headers        map[string]string
		query          string
		mockService    core.MarketDataService
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "missing user header",
			headers:        map[string]string{},
			query:          "?symbol=AAPL",
			mockService:    &mocks.MockMarketDataService{},
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       "missing user authentication context",
		},
		{
			name: "missing bearer token",
			headers: map[string]string{
				handler_adapters.USER_ID_HEADER_KEY: "user1",
				"Authorization":                     "abc",
			},
			query:          "?symbol=AAPL",
			mockService:    &mocks.MockMarketDataService{},
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       "missing authorization token",
		},
		{
			name: "missing symbol query",
			headers: map[string]string{
				handler_adapters.USER_ID_HEADER_KEY: "user1",
				"Authorization":                     "Bearer token",
			},
			query:          "",
			mockService:    &mocks.MockMarketDataService{},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "missing 'symbol' query parameter",
		},
		{
			name: "service returns error",
			headers: map[string]string{
				handler_adapters.USER_ID_HEADER_KEY: "user1",
				"Authorization":                     "Bearer token",
			},
			query:          "?symbol=AAPL",
			mockService:    &mocks.MockMarketDataService{Err: errors.New("db error")},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "failed to get instrument",
		},
		{
			name: "success",
			headers: map[string]string{
				handler_adapters.USER_ID_HEADER_KEY: "user1",
				"Authorization":                     "Bearer token",
			},
			query: "?symbol=AAPL",
			mockService: &mocks.MockMarketDataService{
				Instrument: &models.Instrument{
					Symbol: "AAPL",
					Name:   "Apple Inc.",
				},
			},
			wantStatusCode: http.StatusOK,
			wantBody:       "Apple Inc.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := handler_adapters.MarketDataHandler{
				Service: tt.mockService,
			}
			req := httptest.NewRequest(http.MethodGet, "/price"+tt.query, nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			rr := httptest.NewRecorder()

			handler.GetInstrument(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Errorf("expected status %d, got %d", tt.wantStatusCode, rr.Code)
			}
			if !strings.Contains(rr.Body.String(), tt.wantBody) {
				t.Errorf("expected body to contain %q, got %s", tt.wantBody, rr.Body.String())
			}
		})
	}
}
