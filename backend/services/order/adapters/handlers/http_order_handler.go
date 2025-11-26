package handler_adapters

import (
	"brokerx/order-service/common"
	"brokerx/order-service/core"
	"brokerx/order-service/models"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
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
	if err := validate.RegisterValidation("limitprice", isLimitPriceValid); err != nil {
		slog.Error("could not add order custom validation rule", "error", err)
	}
	return &OrderHandler{OrderService: orderService, ComplianceService: complianceService}
}

func (handler *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	log := common.FromContext(r.Context())

	userIDStr := r.Header.Get(common.HeaderKeyUserId)
	if userIDStr == "" {
		msg := "missing user authentication context"
		log.Warn(msg)
		common.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: msg})
		return
	}
	userId, err := strconv.Atoi(userIDStr)
	if err != nil {
		msg := "invalid user authentication context"
		log.Warn(msg, "error", err)
		common.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: msg})
	}

	jwt := r.Header.Get(common.HeaderKeyAuth)
    if !strings.HasPrefix(jwt, common.AuthHeaderBearerPrefix) {
		msg := "missing authorization token"
		log.Warn(msg)
		common.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: msg})
        return
    }

	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		msg := "invalid JSON format"
		log.Warn(msg, "error", err)
		common.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
		return
	}
	order.UserID = userId
	order.RemainingQuantity = order.Quantity

	if err := validate.Struct(order); err != nil {
		msg := "missing or invalid fields"
		log.Warn(msg, "error", err)
		common.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
		return
	}

	ctx := context.WithValue(r.Context(), common.CtxKeyJWT, jwt)
	ctx = context.WithValue(ctx, common.CtxKeyUserId, userIDStr)

	// if err := handler.ComplianceService.VerifyOrderCompliance(ctx, &order); err != nil {
	// 	status := http.StatusInternalServerError
	// 	if errors.Is(err, common.ErrBusinessRuleViolation) {status = http.StatusBadRequest}
	// 	if errors.Is(err, common.ErrDependencyFailure) {status = http.StatusBadGateway}
	// 	log.Warn(err.Error())
	// 	common.WriteJSON(w, status, ErrorResponse{ErrorMessage: err.Error()})
	// 	return
	// }

	if err := handler.OrderService.PlaceOrder(ctx, &order); err != nil {
		msg := "failed to place order"
		log.Warn(msg, "error", err)
		common.WriteJSON(w, http.StatusInternalServerError, ErrorResponse{ErrorMessage: msg})
		return
	}

	common.WriteJSON(w, http.StatusCreated, OrderPlacedResponse{
		Message: "order placed successfully",
		Symbol:  order.Symbol,
	})
}

func (handler *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	log := common.FromContext(r.Context())

	userIDStr := r.Header.Get(common.HeaderKeyUserId)
	if userIDStr == "" {
		msg := "missing user authentication context"
		log.Warn(msg)
		common.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: msg})
		return
	}
	userId, err := strconv.Atoi(userIDStr)
	if err != nil {
		msg := "invalid user authentication context"
		log.Warn(msg, "error", err)
		common.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: msg})
	}

	jwt := r.Header.Get(common.HeaderKeyAuth)
    if !strings.HasPrefix(jwt, common.AuthHeaderBearerPrefix) {
		msg := "missing authorization token"
		log.Warn(msg)
		common.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: msg})
        return
    }

	orders, err := handler.OrderService.GetOrdersForUser(r.Context(), userId)
	if err != nil {
		msg := "failed to fetch orders"
		log.Warn(msg, "error", err)
		common.WriteJSON(w, http.StatusInternalServerError, ErrorResponse{ErrorMessage: msg})
		return
	}

	common.WriteJSON(w, http.StatusOK, OrdersResponse{Orders: orders})
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
