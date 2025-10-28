package mocks

import "brokerx/matching-service/models"

type MockExecQueue struct {
	Records []*models.ExecutionRecord
	Err     error
}

func (m *MockExecQueue) EnqueueExecutionRecords(records []*models.ExecutionRecord) error {
	if m.Err != nil {
		return m.Err
	}
	return nil
}

func (m *MockExecQueue) DequeueExecutionRecords(batchSize int) ([]*models.ExecutionRecord, error) {
	if m.Err != nil {
		return []*models.ExecutionRecord{}, m.Err
	}
	return m.Records, nil
}
