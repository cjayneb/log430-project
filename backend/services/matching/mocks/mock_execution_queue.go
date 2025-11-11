package mocks

import (
	"brokerx/matching-service/models"
	"context"
)

type MockExecQueue struct {
	Records []*models.ExecutionRecord
	Err     error
}

func (m *MockExecQueue) EnqueueExecutionRecords(ctx context.Context, records []*models.ExecutionRecord) error {
	if m.Err != nil {
		return m.Err
	}
	return nil
}
