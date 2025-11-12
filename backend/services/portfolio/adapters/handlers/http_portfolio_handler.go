package handler_adapters

import (
	"brokerx/portfolio-service/core"
	"brokerx/portfolio-service/models"
	"brokerx/portfolio-service/util"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
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
	log := util.FromContext(r.Context())

	userIDStr := r.Header.Get(USER_ID_HEADER_KEY)
	if userIDStr == "" {
		msg := "missing user authentication context"
		log.Warn(msg)
		util.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: msg})
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		msg := "invalid 'user_id' parameter"
		log.Warn(msg)
		util.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
		return
	}

	wallet, err := handler.Service.GetWallet(r.Context(), userID)
	if err != nil {
		msg := "failed to get wallet"
		log.Warn(msg, "error", err)
		util.WriteJSON(w, http.StatusInternalServerError, ErrorResponse{ErrorMessage: msg})
		return
	}

	util.WriteJSON(w, http.StatusOK, WalletResponse{Wallet: *wallet})
}

func (handler *PortfolioHandler) FundWallet(w http.ResponseWriter, r *http.Request) {
	log := util.FromContext(r.Context())

	userIDStr := r.Header.Get(USER_ID_HEADER_KEY)
	if userIDStr == "" {
		msg := "missing user authentication context"
		log.Warn(msg)
		util.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: msg})
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		msg := "invalid 'user_id' parameter"
		log.Warn(msg)
		util.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
		return
	}

	var addFundsReq AddFundsRequest
	if err := json.NewDecoder(r.Body).Decode(&addFundsReq); err != nil {
		msg := "invalid JSON format"
		log.Warn(msg, "error", err)
		util.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
		return
	}
	if err := validate.Struct(addFundsReq); err != nil {
		msg := "missing or invalid fields"
		log.Warn(msg, "error", err)
		util.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
		return
	}

	if err = handler.Service.FundWallet(r.Context(), userID, addFundsReq.Amount); err != nil {
		msg := "failed to fund wallet"
		log.Warn(msg, "error", err)
		util.WriteJSON(w, http.StatusInternalServerError, ErrorResponse{ErrorMessage: msg})
		return
	}

	wallet, err := handler.Service.GetWallet(r.Context(), userID)
	if err != nil {
		msg := "failed to get wallet after adding funds"
		log.Warn(msg, "error", err)
		util.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
		return
	}

	util.WriteJSON(w, http.StatusOK, WalletResponse{Wallet: *wallet})
}

func (handler *PortfolioHandler) FetchPositions(w http.ResponseWriter, r *http.Request) {
	log := util.FromContext(r.Context())

	userIDStr := r.Header.Get(USER_ID_HEADER_KEY)
	if userIDStr == "" {
		msg := "missing user authentication context"
		log.Warn(msg)
		util.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: msg})
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		msg := "invalid 'user_id' parameter"
		log.Warn(msg)
		util.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
		return
	}

	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		msg := "missing 'symbol' query parameter"
		log.Warn(msg)
		util.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
		return
	}

	positions, err := handler.Service.FetchPositions(r.Context(), userID, symbol)
	if err != nil {
		msg := "failed to fetch positions"
		log.Warn(msg, "error", err)
		util.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
		return
	}

	util.WriteJSON(w, http.StatusOK, PositionsResponse{Positions: positions})
}
