package mocks

import (
	"brokerx/matching-service/models"
	"brokerx/matching-service/ports"
)

type MockOrderBook struct {
	Order  models.Order
	Orders []*models.Order
	Err    error
}

func (m *MockOrderBook) EnqueueOrders(orders []*models.Order) error {
	if m.Err != nil {
		return m.Err
	}
	return nil
}

func (m *MockOrderBook) DequeueOrders(batchSize int) ([]*models.Order, error) {
	if m.Err != nil {
		return []*models.Order{}, m.Err
	}
	return m.Orders, nil
}

func (m *MockOrderBook) GetById(orderId int) (models.Order, error) {
	if m.Err != nil {
		return models.Order{}, m.Err
	}
	return m.Order, nil
}

func (m *MockOrderBook) FindMatchesLimit(symbol string, action string, unitPrice float64, batchSize int) ([]*models.Order, error) {
	if m.Err != nil {
		return []*models.Order{}, m.Err
	}
	return m.Orders, nil
}

func (m *MockOrderBook) FindMatchesMarket(symbol string, action string, batchSize int) ([]*models.Order, error) {
	if m.Err != nil {
		return []*models.Order{}, m.Err
	}
	return m.Orders, nil
}

func (m *MockOrderBook) Insert(order *models.Order) error {
	return m.Err
}

func (m *MockOrderBook) Return(orders []*models.Order) error {
	return m.Err
}

func (m *MockOrderBook) MarkDirty(orderID int) error {
	return m.Err
}

var _ ports.OrderBook = (*MockOrderBook)(nil)
