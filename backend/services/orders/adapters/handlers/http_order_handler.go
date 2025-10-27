package handler_adapters

import (
	"brokerx/order-service/common"
	"brokerx/order-service/core"
	"brokerx/order-service/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

type ErrorResponse struct {
	ErrorMessage string `json:"errorMessage"`
}

type OrderPlacedResponse struct {
	Message string `json:"message"`
	Symbol  string `json:"symbol"`
}

type OrdersResponse struct {
	Orders []*models.Order `json:"orders"`
}

type OrderHandler struct {
	OrderService core.OrderService
	ComplianceService core.ComplianceService
}

var validate = validator.New()

func NewOrderHandler(orderService core.OrderService, complianceService core.ComplianceService) *OrderHandler {
	validate.RegisterValidation("limitprice", isLimitPriceValid)
	return &OrderHandler{OrderService: orderService, ComplianceService: complianceService}
}

func (handler *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get(common.HeaderKeyUserId)
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: "missing user authentication context"})
		return
	}

	jwt := r.Header.Get(common.HeaderKeyAuth)
    if !strings.HasPrefix(jwt, common.AuthHeaderBearerPrefix) {
        writeJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: "missing authorization token"})
        return
    }

	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		log.Warnf("Invalid JSON: %v", err)
		writeJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: "invalid JSON format"})
		return
	}
	order.UserID = userID
	order.RemainingQuantity = order.Quantity

	if err := validate.Struct(order); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: fmt.Sprintf("missing or invalid fields: %v", err)})
		return
	}

	ctx := context.WithValue(r.Context(), common.CtxKeyJWT, jwt)

	if err := handler.ComplianceService.VerifyOrderCompliance(ctx, &order); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, common.ErrBusinessRuleViolation) {status = http.StatusBadRequest}
		if errors.Is(err, common.ErrDependencyFailure) {status = http.StatusBadGateway}
		writeJSON(w, status, ErrorResponse{ErrorMessage: err.Error()})
		return
	}

	if err := handler.OrderService.PlaceOrder(ctx, &order); err != nil {
		log.Errorf("Failed to place order : %v", err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{ErrorMessage: "failed to place order"})
		return
	}

	writeJSON(w, http.StatusCreated, OrderPlacedResponse{
		Message: "order placed successfully",
		Symbol:  order.Symbol,
	})
}

func (handler *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get(common.HeaderKeyUserId)
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: "missing user authentication context"})
		return
	}

	orders, err := handler.OrderService.GetOrdersForUser(userID)
	if err != nil {
		log.Errorf("Failed to fetch orders for user %v: %v", userID, err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{ErrorMessage: "failed to fetch orders"})
		return
	}

	writeJSON(w, http.StatusOK, OrdersResponse{Orders: orders})
}

func isLimitPriceValid(fl validator.FieldLevel) bool {
	order, ok := fl.Top().Interface().(models.Order)
	if !ok {
		return false
	}

	if order.Type == "limit" {
		return order.UnitPrice > 0
	}

	return true
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("error when encoding JSON response : %v", err), http.StatusInternalServerError)
	}
}
