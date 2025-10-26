package handler_adapters

import (
	"brokerx/matching-service/core"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

const USER_ID_HEADER_KEY string = "X-User-Id"

type SubmitOrderRequest struct {
	OrderId int `json:"order_id" validate:"required"`
}

type OrderSubmittedResponse struct {
	Message string `json:"message"`
	OrderId int    `json:"order_id"`
}

type ErrorResponse struct {
	ErrorMessage string `json:"errorMessage"`
}

type MatchingHandler struct {
	MatchingEngine core.MatchingEngine
}

var validate = validator.New()

func (handler *MatchingHandler) SubmitOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get(USER_ID_HEADER_KEY)
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: "missing user authentication context"})
		return
	}

	var submitReq SubmitOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&submitReq); err != nil {
		log.Warnf("Invalid JSON: %v", err)
		writeJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: "invalid JSON format"})
		return
	}
	if err := validate.Struct(submitReq); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: fmt.Sprintf("missing or invalid fields: %v", err)})
		return
	}

	if err := handler.MatchingEngine.SubmitOrder(submitReq.OrderId); err != nil {
		log.Errorf("Failed to submit order : %v", err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{ErrorMessage: "failed to submit order"})
		return
	}

	log.Infof("Order #%d submitted to matching engine", submitReq.OrderId)
	writeJSON(w, http.StatusAccepted, OrderSubmittedResponse{
		Message: fmt.Sprintf("order #%d submitted to the matching engine", submitReq.OrderId),
		OrderId: submitReq.OrderId,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("error when encoding JSON response : %v", err), http.StatusInternalServerError)
	}
}
