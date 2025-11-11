package mocks

import (
	"brokerx/matching-service/models"
	"context"
)

type MockMatchingEngine struct {
	Err error
}

func (m *MockMatchingEngine) QueueOrder(ctx context.Context, order *models.Order) error {
	if m.Err != nil {
		return m.Err
	}
	return nil
}
