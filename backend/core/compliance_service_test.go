package core

import (
	"brokerx/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockWalletRepo struct {
	mock.Mock
}

func (m *MockWalletRepo) FindByUserId(userId string) (*models.Wallet, error) {
	args := m.Called(userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

type MockPositionsRepo struct {
	mock.Mock
}

func (m *MockPositionsRepo) FindByUserIdAndSymbol(userId string, symbol string) ([]*models.Position, error) {
	args := m.Called(userId, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Position), args.Error(1)
}

func makeWallet(order *models.Order) *models.Wallet {
	return &models.Wallet{
		UserId: order.UserID,
		AvailableFunds: (order.UnitPrice * float64(order.Quantity)) + 100,
		OnHoldFunds: 100.0,
	}
}

func makePositions(order *models.Order) [] *models.Position {
	return []*models.Position{
		{UserId: order.UserID, Symbol: order.Symbol, Quantity: order.Quantity + 1},
	}
}

// ---------------------------
// Test Suite
// ---------------------------

type ComplianceServiceTestSuite struct {
	suite.Suite
	walletRepo    *MockWalletRepo
	positionRepo *MockPositionsRepo
	service *ComplianceService
}

func (s *ComplianceServiceTestSuite) SetupTest() {
	s.walletRepo = new(MockWalletRepo)
	s.positionRepo = new(MockPositionsRepo)
	s.service = &ComplianceService{WalletRepo: s.walletRepo, PositionRepo: s.positionRepo}
}

// ---------------------------
// Tests
// ---------------------------

func (s *ComplianceServiceTestSuite) TestVerifyBuyOrderSuccess() {
	order := makeOrder()
	wallet := makeWallet(order)
	s.walletRepo.On("FindByUserId", order.UserID).Return(wallet, nil)

	err := s.service.VerifyOrderCompliance(order)

	s.Require().NoError(err)
}

func (s *ComplianceServiceTestSuite) TestVerifyBuyOrderNonCompliance() {
	order := makeOrder()
	wallet := makeWallet(order)
	wallet.AvailableFunds = (order.UnitPrice * float64(order.Quantity)) - 5
	s.walletRepo.On("FindByUserId", order.UserID).Return(wallet, nil)

	err := s.service.VerifyOrderCompliance(order)

	s.EqualError(err, "not enough available funds")
}

func (s *ComplianceServiceTestSuite) TestVerifyBuyOrderFailure() {
	order := makeOrder()
	s.walletRepo.On("FindByUserId", order.UserID).Return(nil, assert.AnError)

	err := s.service.VerifyOrderCompliance(order)

	s.Error(err)
}

func (s *ComplianceServiceTestSuite) TestVerifySellOrderSuccess() {
	order := makeOrder()
	order.Action = "sell"
	positions := makePositions(order)
	s.positionRepo.On("FindByUserIdAndSymbol", order.UserID, order.Symbol).Return(positions, nil)

	err := s.service.VerifyOrderCompliance(order)

	s.Require().NoError(err)
}

func (s *ComplianceServiceTestSuite) TestVerifySellOrderNonCompliance() {
	order := makeOrder()
	order.Action = "sell"
	s.positionRepo.On("FindByUserIdAndSymbol", order.UserID, order.Symbol).Return(make([]*models.Position, 0), nil)

	err := s.service.VerifyOrderCompliance(order)

	s.EqualError(err, "not enough owned stocks")
}

func (s *ComplianceServiceTestSuite) TestVerifySellOrderFailure() {
	order := makeOrder()
	order.Action = "sell"
	s.positionRepo.On("FindByUserIdAndSymbol", order.UserID, order.Symbol).Return(nil, assert.AnError)

	err := s.service.VerifyOrderCompliance(order)

	s.Error(err)
}

// ---------------------------
// Run the suite
// ---------------------------
func TestComplianceServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ComplianceServiceTestSuite))
}
