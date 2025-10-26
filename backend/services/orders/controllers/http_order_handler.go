package controllers

import (
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

const USER_ID_HEADER_KEY string = "X-User-Id"

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
	Service ports.OrderService
}

var validate = validator.New()

func NewOrderHandler(service ports.OrderService) *OrderHandler {
	validate.RegisterValidation("limitprice", isLimitPriceValid)
	return &OrderHandler{Service: service}
}

func (handler *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get(USER_ID_HEADER_KEY)
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: "missing user authentication context"})
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

	if err := handler.Service.PlaceOrder(&order); err != nil {
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
	userID := r.Header.Get(USER_ID_HEADER_KEY)
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: "missing user authentication context"})
		return
	}

	orders, err := handler.Service.GetOrdersForUser(userID)
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
