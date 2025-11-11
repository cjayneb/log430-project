package handler_adapters

import (
	"brokerx/market-data-service/core"
	"brokerx/market-data-service/models"
	"brokerx/market-data-service/util"
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
	log := util.FromContext(r.Context())

	if r.Header.Get(USER_ID_HEADER_KEY) == "" {
		msg := "missing user authentication context"
		log.Warn(msg)
		util.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: msg})
		return
	}

	jwt := r.Header.Get("Authorization")
	if !strings.HasPrefix(jwt, "Bearer ") {
		msg := "missing authorization token"
		log.Warn(msg)
		util.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: msg})
		return
	}

	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		msg := "missing 'symbol' query parameter"
		log.Warn(msg)
		util.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
		return
	}

	price, err := handler.Service.GetCurrentStockPriceBySymbol(r.Context(), symbol)
	if err != nil {
		msg := "failed to get price"
		log.Warn(msg, "error", err)
		util.WriteJSON(w, http.StatusInternalServerError, ErrorResponse{ErrorMessage: msg})
		return
	}

	util.WriteJSON(w, http.StatusOK, StockPriceResponse{Symbol: strings.ToUpper(symbol), Price: price})
}

func (handler *MarketDataHandler) GetInstrument(w http.ResponseWriter, r *http.Request) {
	log := util.FromContext(r.Context())

	if r.Header.Get(USER_ID_HEADER_KEY) == "" {
		msg := "missing user authentication context"
		log.Warn(msg)
		util.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: msg})
		return
	}

	jwt := r.Header.Get("Authorization")
	if !strings.HasPrefix(jwt, "Bearer ") {
		msg := "missing authorization token"
		log.Warn(msg)
		util.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: msg})
		return
	}

	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		msg := "missing 'symbol' query parameter"
		log.Warn(msg)
		util.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
		return
	}

	instrument, err := handler.Service.GetInstrumentBySymbol(r.Context(), symbol)
	if err != nil {
		msg := "failed to get instrument"
		log.Warn(msg, "error", err)
		util.WriteJSON(w, http.StatusInternalServerError, ErrorResponse{ErrorMessage: msg})
		return
	}

	util.WriteJSON(w, http.StatusOK, InstrumentResponse{Instrument: *instrument})
}
