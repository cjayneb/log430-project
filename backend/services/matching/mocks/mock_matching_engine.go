package mocks

import "brokerx/matching-service/models"

type MockMatchingEngine struct {
	Err error
}

func (m *MockMatchingEngine) QueueOrder(order *models.Order) error {
	if m.Err != nil {
		return m.Err
	}
	return nil
}
