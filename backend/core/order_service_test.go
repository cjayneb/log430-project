package core

import (
	"brokerx/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockOrderRepo struct {
	mock.Mock
}

func (m *MockOrderRepo) CreateOrder(order *models.Order) (int, error) {
	args := m.Called(order)
	if args.Get(0) == nil {
		return 0, args.Error(1)
	}
	return args.Get(0).(int), args.Error(1)
}

type MockComplianceService struct {
	mock.Mock
}

func (m *MockComplianceService) VerifyOrderCompliance(order *models.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func makeOrder() *models.Order {
	return &models.Order{
		UserID: uuid.New().String(),
		Symbol: "AAPL",
		Type:   "market",
		Action: "buy",
		Quantity:  10,
		UnitPrice: 150.00,
		Timing:    "day",
		Status:    "open",
	}
}

// ---------------------------
// Test Suite
// ---------------------------

type OrderServiceTestSuite struct {
	suite.Suite
	repo    *MockOrderRepo
	complianceService *MockComplianceService
	service *OrderService
}

func (s *OrderServiceTestSuite) SetupTest() {
	s.repo = new(MockOrderRepo)
	s.complianceService = new(MockComplianceService)
	s.service = &OrderService{Repo: s.repo, ComplianceService: s.complianceService}
}

// ---------------------------
// Tests
// ---------------------------

func (s *OrderServiceTestSuite) TestPlaceOrderSuccess() {
	order := makeOrder()
	s.complianceService.On("VerifyOrderCompliance", order).Return(nil)
	s.repo.On("CreateOrder", order).Return(1, nil)

	err := s.service.PlaceOrder(order)

	s.Require().NoError(err)
}

func (s *OrderServiceTestSuite) TestPlaceOrderNonCompliance() {
	order := makeOrder()
	s.complianceService.On("VerifyOrderCompliance", order).Return(assert.AnError)

	err := s.service.PlaceOrder(order)

	s.Error(err)
}

func (s *OrderServiceTestSuite) TestPlaceOrderFailure() {
	order := makeOrder()
	s.complianceService.On("VerifyOrderCompliance", order).Return(nil)
	s.repo.On("CreateOrder", order).Return(0, assert.AnError)

	err := s.service.PlaceOrder(order)

	s.Error(err)
}

// ---------------------------
// Run the suite
// ---------------------------
func TestOrderServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OrderServiceTestSuite))
}
