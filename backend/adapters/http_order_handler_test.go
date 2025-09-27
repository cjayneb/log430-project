package adapters

import (
	"brokerx/models"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var PLACE_ORDER_ENDPOINT string = "/order/place"

type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) PlaceOrder(order *models.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

// ---------------------------
// Test Suite
// ---------------------------

type HttpOrderHandlerTestSuite struct {
	suite.Suite
	mockService *MockOrderService
	handler *OrderHandler
	UserID string
	Symbol string
	Type   string
	Action string
	Quantity  int
	UnitPrice float64
	Timing    string
	Status    string
	RequestString string
}

func (s *HttpOrderHandlerTestSuite) SetupTest() {
	s.mockService = new(MockOrderService)
	s.handler = &OrderHandler{Service: s.mockService}
	s.UserID = uuid.New().String()
	s.Symbol = "AAPL"
	s.Type = "market"
	s.Action = "buy"
	s.Quantity = 10
	s.UnitPrice = 150.00
	s.Timing = "day"
	s.Status = "open"
	s.RequestString = fmt.Sprintf("user_id=%s&symbol=%s&type=%s&action=%s&quantity=%d&unit_price=%.2f&timing=%s&status=%s",
		s.UserID, s.Symbol, s.Type, s.Action, s.Quantity, s.UnitPrice, s.Timing, s.Status)
}

func (s *HttpOrderHandlerTestSuite) TestPlaceOrderSuccess() {
	s.SetupTest()
	s.mockService.On("PlaceOrder", mock.AnythingOfType("*models.Order")).Return(nil)

	req := httptest.NewRequest(http.MethodPost, PLACE_ORDER_ENDPOINT, bytes.NewBufferString(s.RequestString))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), USER_ID_KEY, s.UserID))
	w := httptest.NewRecorder()

	s.handler.PlaceOrder(w, req)
	res := w.Result()
	defer res.Body.Close()

	s.Equal(http.StatusCreated, res.StatusCode)
}

func (s *HttpOrderHandlerTestSuite) TestPlaceOrderBadRequest() {
	s.SetupTest()

	req := httptest.NewRequest(http.MethodPost, PLACE_ORDER_ENDPOINT, bytes.NewBufferString("quantity=ten"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), USER_ID_KEY, s.UserID))
	w := httptest.NewRecorder()

	s.handler.PlaceOrder(w, req)
	res := w.Result()
	defer res.Body.Close()

	s.Equal(http.StatusBadRequest, res.StatusCode)
}

func (s *HttpOrderHandlerTestSuite) TestPlaceOrderInternalError() {
	s.SetupTest()
	s.mockService.On("PlaceOrder", mock.AnythingOfType("*models.Order")).Return(assert.AnError)

	req := httptest.NewRequest(http.MethodPost, PLACE_ORDER_ENDPOINT, bytes.NewBufferString(s.RequestString))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), USER_ID_KEY, s.UserID))
	w := httptest.NewRecorder()

	s.handler.PlaceOrder(w, req)
	res := w.Result()
	defer res.Body.Close()

	s.Equal(http.StatusInternalServerError, res.StatusCode)
}

// ---------------------------
// Run the suite
// ---------------------------
func TestHttpOrderHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HttpOrderHandlerTestSuite))
}
