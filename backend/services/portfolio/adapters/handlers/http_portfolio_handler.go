package handler_adapters

import (
	"brokerx/portfolio-service/core"
	"brokerx/portfolio-service/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

const USER_ID_HEADER_KEY string = "X-User-Id"

type AddFundsRequest struct {
	Amount float64 `json:"amount" validate:"gt=0"`
}

type WalletResponse struct {
	Wallet models.Wallet `json:"wallet"`
}

type PositionsResponse struct {
	Positions []*models.Position `json:"positions"`
}

type ErrorResponse struct {
	ErrorMessage string `json:"errorMessage"`
}

type PortfolioHandler struct {
	Service core.PortfolioService
}

var validate = validator.New()

func (handler *PortfolioHandler) GetWallet(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get(USER_ID_HEADER_KEY)
	if userIDStr == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: "missing user authentication context"})
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: "invalid 'user_id' parameter"})
		return
	}

	wallet, err := handler.Service.GetWallet(userID)
	if err != nil {
		log.Errorf("Failed to get wallet for user %d : %v", userID, err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{ErrorMessage: "failed to get wallet"})
		return
	}

	writeJSON(w, http.StatusOK, WalletResponse{Wallet: *wallet})
}

func (handler *PortfolioHandler) FundWallet(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get(USER_ID_HEADER_KEY)
	if userIDStr == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: "missing user authentication context"})
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: "invalid 'user_id' parameter"})
		return
	}

	var addFundsReq AddFundsRequest
	if err := json.NewDecoder(r.Body).Decode(&addFundsReq); err != nil {
		log.Warnf("Invalid JSON: %v", err)
		writeJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: "invalid JSON format"})
		return
	}
	if err := validate.Struct(addFundsReq); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: fmt.Sprintf("missing or invalid fields: %v", err)})
		return
	}

	if err = handler.Service.FundWallet(userID, addFundsReq.Amount); err != nil {
		log.Errorf("Failed to fund wallet for user %d : %v", userID, err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{ErrorMessage: "failed to fund wallet"})
		return
	}

	wallet, err := handler.Service.GetWallet(userID)
	if err != nil {
		log.Errorf("Failed to get wallet for user %d : %v", userID, err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{ErrorMessage: "failed to get wallet"})
		return
	}

	writeJSON(w, http.StatusOK, WalletResponse{Wallet: *wallet})
}

func (handler *PortfolioHandler) FetchPositions(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get(USER_ID_HEADER_KEY)
	if userIDStr == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: "missing user authentication context"})
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: "invalid 'user_id' parameter"})
		return
	}

	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: "missing 'symbol' query parameter"})
		return
	}

	positions, err := handler.Service.FetchPositions(userID, symbol)
	if err != nil {
		log.Errorf("Failed to fetch positions for user %d and symbol %s : %v", userID, symbol, err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{ErrorMessage: "failed to fetch positions"})
		return
	}

	writeJSON(w, http.StatusOK, PositionsResponse{Positions: positions})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("error when encoding JSON response : %v", err), http.StatusInternalServerError)
	}
}
