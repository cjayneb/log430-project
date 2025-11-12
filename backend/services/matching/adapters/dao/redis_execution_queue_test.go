package dao_adapters_test

import (
	dao_adapters "brokerx/matching-service/adapters/dao"
	"brokerx/matching-service/mocks"
	"brokerx/matching-service/models"
	"context"
	"fmt"
	"os"
	"testing"
)

var redisClientMock, s = mocks.GetRedisClientMock()
var queue = dao_adapters.RedisExecutionQueue{Rdb: redisClientMock}

var exec = models.ExecutionRecord{
	BuyOrderID:  1,
	SellOrderID: 2,
	Symbol:      "AAPL",
	Price:       123.0,
	Quantity:    23,
}

var baseExecs = []*models.ExecutionRecord{&exec, &exec}

var execsToAdd = []*models.ExecutionRecord{&exec}

func TestMain(m *testing.M) {
	code := m.Run()

	s.Close()

	os.Exit(code)
}

func setup(wantErr bool) {
	if wantErr {
		s.Close()
	} else {
		redisClientMock, s = mocks.GetRedisClientMock()
		queue = dao_adapters.RedisExecutionQueue{Rdb: redisClientMock}
		_ = queue.EnqueueExecutionRecords(ctx, baseExecs)
	}
}

func checkExecQueueLength(expected int64) {
	length, _ := redisClientMock.LLen(context.Background(), dao_adapters.EXECUTION_RECORDS_PERSISTANCE_QUEUE).Result()
	if length != expected {
		panic(fmt.Sprintf("EnqueueExecutionRecords() = %v, want %v", length, expected))
	}
}

func TestRedisExecutionQueue_EnqueueExecutionRecords(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		records []*models.ExecutionRecord
		want    int64
		wantErr bool
	}{
		{
			name:    "when enqueues empty list of records then no records except already present ones",
			records: []*models.ExecutionRecord{},
			want:    2,
			wantErr: false,
		},
		{
			name:    "when enqueues list of execution records then records and no error",
			records: execsToAdd,
			want:    3,
			wantErr: false,
		},
		{
			name:    "when redis is down then error",
			records: execsToAdd,
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup(tt.wantErr)

			gotErr := queue.EnqueueExecutionRecords(ctx, tt.records)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("EnqueueExecutionRecords() failed: %v", gotErr)
				}
				checkExecQueueLength(tt.want)
				return
			}
			if tt.wantErr {
				t.Fatal("EnqueueExecutionRecords() succeeded unexpectedly")
			}
			checkExecQueueLength(tt.want)
		})
	}
}
