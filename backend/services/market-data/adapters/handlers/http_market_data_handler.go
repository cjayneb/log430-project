package handler_adapters

import (
	"brokerx/market-data-service/core"
	"brokerx/market-data-service/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const USER_ID_HEADER_KEY string = "X-User-Id"

type StockPriceResponse struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
}

type InstrumentResponse struct {
	Instrument models.Instrument `json:"instrument"`
}

type ErrorResponse struct {
	ErrorMessage string `json:"errorMessage"`
}

type MarketDataHandler struct {
	Service core.MarketDataService
}

func (handler *MarketDataHandler) GetStockPrice(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(USER_ID_HEADER_KEY) == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: "missing user authentication context"})
		return
	}

	jwt := r.Header.Get("Authorization")
	if !strings.HasPrefix(jwt, "Bearer ") {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: "missing authorization token"})
		return
	}

	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: "missing 'symbol' query parameter"})
		return
	}

	price, err := handler.Service.GetCurrentStockPriceBySymbol(symbol)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{ErrorMessage: fmt.Sprintf("failed to get price : %v", err)})
		return
	}

	writeJSON(w, http.StatusOK, StockPriceResponse{Symbol: strings.ToUpper(symbol), Price: price})
}

func (handler *MarketDataHandler) GetInstrument(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(USER_ID_HEADER_KEY) == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: "missing user authentication context"})
		return
	}

	jwt := r.Header.Get("Authorization")
	if !strings.HasPrefix(jwt, "Bearer ") {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: "missing authorization token"})
		return
	}

	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: "missing 'symbol' query parameter"})
		return
	}

	instrument, err := handler.Service.GetInstrumentBySymbol(symbol)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{ErrorMessage: fmt.Sprintf("failed to get instrument : %v", err)})
		return
	}

	writeJSON(w, http.StatusOK, InstrumentResponse{Instrument: *instrument})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("error when encoding JSON response : %v", err), http.StatusInternalServerError)
	}
}
