package handler_adapters

import (
	"brokerx/portfolio-service/core"
	"brokerx/portfolio-service/models"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

const USER_ID_HEADER_KEY string = "X-User-Id"

type WalletRequest struct {
	UserID int `json:"user_id" validate:"required"`
}
type WalletResponse struct {
	Wallet models.Wallet `json:"wallet"`
}

type PositionsRequest struct {
	UserID int    `json:"user_id" validate:"required"`
	Symbol string `json:"symbol" validate:"required"`
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
	userID := r.Header.Get(USER_ID_HEADER_KEY)
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: "missing user authentication context"})
		return
	}

	var walletReq WalletRequest
	if err := json.NewDecoder(r.Body).Decode(&walletReq); err != nil {
		log.Warnf("Invalid JSON: %v", err)
		writeJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: "invalid JSON format"})
		return
	}
	if err := validate.Struct(walletReq); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: fmt.Sprintf("missing or invalid fields: %v", err)})
		return
	}

	wallet, err := handler.Service.GetWallet(walletReq.UserID)
	if err != nil {
		log.Errorf("Failed to get wallet for user %d : %v", walletReq.UserID, err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{ErrorMessage: "failed to get wallet"})
		return
	}

	writeJSON(w, http.StatusOK, WalletResponse{Wallet: *wallet})
}

func (handler *PortfolioHandler) FetchPositions(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get(USER_ID_HEADER_KEY)
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: "missing user authentication context"})
		return
	}

	var positionsReq PositionsRequest
	if err := json.NewDecoder(r.Body).Decode(&positionsReq); err != nil {
		log.Warnf("Invalid JSON: %v", err)
		writeJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: "invalid JSON format"})
		return
	}
	if err := validate.Struct(positionsReq); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: fmt.Sprintf("missing or invalid fields: %v", err)})
		return
	}

	positions, err := handler.Service.FetchPositions(positionsReq.UserID, positionsReq.Symbol)
	if err != nil {
		log.Errorf("Failed to fetch positions for user %d and symbol %s : %v", positionsReq.UserID, positionsReq.Symbol, err)
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
