package mocks

import (
	"brokerx/matching-service/models"
	"brokerx/matching-service/ports"
	"context"
)

type MockOrderBook struct {
	Order  models.Order
	Orders []*models.Order
	Err    error
}

func (m *MockOrderBook) EnqueueOrders(ctx context.Context, orders []*models.Order) error {
	if m.Err != nil {
		return m.Err
	}
	return nil
}

func (m *MockOrderBook) DequeueOrders(ctx context.Context, batchSize int) ([]*models.Order, error) {
	if m.Err != nil {
		return []*models.Order{}, m.Err
	}
	return m.Orders, nil
}

func (m *MockOrderBook) GetById(ctx context.Context, orderId int) (models.Order, error) {
	if m.Err != nil {
		return models.Order{}, m.Err
	}
	return m.Order, nil
}

func (m *MockOrderBook) FindMatches(ctx context.Context, order *models.Order, batchSize int) ([]*models.Order, error) {
	if m.Err != nil {
		return []*models.Order{}, m.Err
	}
	return m.Orders, nil
}

func (m *MockOrderBook) Insert(ctx context.Context, order *models.Order) error {
	return m.Err
}

func (m *MockOrderBook) Return(ctx context.Context, orders []*models.Order) {}

func (m *MockOrderBook) MarkDirty(ctx context.Context, orderID int) error {
	return m.Err
}

var _ ports.OrderBook = (*MockOrderBook)(nil)
