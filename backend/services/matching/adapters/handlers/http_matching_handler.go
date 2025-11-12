package handler_adapters

import (
	"brokerx/matching-service/core"
	"brokerx/matching-service/models"
	"brokerx/matching-service/util"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

const USER_ID_HEADER_KEY string = "X-User-Id"

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
	log := util.FromContext(r.Context())

	userID := r.Header.Get(USER_ID_HEADER_KEY)
	if userID == "" {
		msg := "missing user authentication context"
		log.Warn(msg)
		util.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: msg})
		return
	}

	var submitReq models.Order
	if err := json.NewDecoder(r.Body).Decode(&submitReq); err != nil {
		msg := "invalid JSON format"
		log.Warn(msg, "error", err)
		util.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
		return
	}
	if err := validate.Struct(submitReq); err != nil {
		msg := "missing or invalid fields"
		log.Warn(msg, "error", err)
		util.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
		return
	}

	if err := handler.MatchingEngine.QueueOrder(r.Context(), &submitReq); err != nil {
		msg := "failed to queue order"
		log.Warn(msg, "error", err)
		util.WriteJSON(w, http.StatusInternalServerError, ErrorResponse{ErrorMessage: msg})
		return
	}

	log.Info("Order submitted to matching engine", "orderId", submitReq.ID)
	util.WriteJSON(w, http.StatusAccepted, OrderSubmittedResponse{
		Message: "order submitted to the matching engine",
		OrderId: submitReq.ID,
	})
}
