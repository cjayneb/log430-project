package handler_adapters_test

import (
	handler_adapters "brokerx/matching-service/adapters/handlers"
	"brokerx/matching-service/core"
	"brokerx/matching-service/mocks"
	"brokerx/matching-service/models"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var invalidOrder = models.Order{
	ID:                1,
	UserID:            1,
	Type:              "market",
	Action:            "sell",
	RemainingQuantity: 1,
	Quantity:          1,
	UnitPrice:         123.0,
	Timing:            "ioc",
}
var validOrder = invalidOrder

func TestMatchingHandler_SubmitOrder(t *testing.T) {
	validOrder.Symbol = "AAPL"

	tests := []struct {
		name           string
		headers        map[string]string
		query          models.Order
		mockService    core.MatchingEngine
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "missing user header",
			headers:        map[string]string{},
			query:          models.Order{},
			mockService:    &mocks.MockMatchingEngine{},
			wantStatusCode: http.StatusUnauthorized,
			wantBody:       "missing user authentication context",
		},
		{
			name: "invalid order",
			headers: map[string]string{
				handler_adapters.USER_ID_HEADER_KEY: "user1",
				"Authorization":                     "Bearer token",
			},
			query:          invalidOrder,
			mockService:    &mocks.MockMatchingEngine{},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `missing or invalid fields: Key: 'Order.Symbol'`,
		},
		{
			name: "service returns error",
			headers: map[string]string{
				handler_adapters.USER_ID_HEADER_KEY: "user1",
				"Authorization":                     "Bearer token",
			},
			query:          validOrder,
			mockService:    &mocks.MockMatchingEngine{Err: errors.New("redis error")},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "failed to queue order",
		},
		{
			name: "success",
			headers: map[string]string{
				handler_adapters.USER_ID_HEADER_KEY: "user1",
				"Authorization":                     "Bearer token",
			},
			query:          validOrder,
			mockService:    &mocks.MockMatchingEngine{},
			wantStatusCode: http.StatusAccepted,
			wantBody:       "order submitted",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := handler_adapters.MatchingHandler{
				MatchingEngine: tt.mockService,
			}
			jsonBody, _ := json.Marshal(tt.query)
			req := httptest.NewRequest(http.MethodPost, "/matching", bytes.NewBuffer(jsonBody))
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			rr := httptest.NewRecorder()

			handler.SubmitOrder(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Errorf("expected status %d, got %d", tt.wantStatusCode, rr.Code)
			}
			if !strings.Contains(rr.Body.String(), tt.wantBody) {
				t.Errorf("expected body to contain %q, got %s", tt.wantBody, rr.Body.String())
			}
		})
	}
}
