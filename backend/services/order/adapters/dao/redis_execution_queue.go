package dao_adapters

import (
	"brokerx/order-service/common"
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

const EXECUTION_RECORDS_PERSISTANCE_QUEUE = "execrecordspersistancequeue"

type RedisExecutionQueue struct {
	Rdb *redis.Client
}

func (queue *RedisExecutionQueue) DequeueExecutionRecords(ctx context.Context, batchSize int) ([]*models.ExecutionRecord, error) {
	log := common.FromContext(ctx)

	var records []*models.ExecutionRecord
	for i := 0; i < batchSize; i++ {
		val, err := queue.Rdb.RPop(ctx, EXECUTION_RECORDS_PERSISTANCE_QUEUE).Result()
		if err == redis.Nil {
			break // queue is empty
		}
		if err != nil {
			log.Error("error dequeuing records from execution records queue", "error", err)
			return nil, err
		}
		var record models.ExecutionRecord
		if err := json.Unmarshal([]byte(val), &record); err != nil {
			log.Error("error unmarshaling dequeued execution record", "error", err)
			continue
		}
		records = append(records, &record)
	}
	return records, nil
}

var _ ports.ExecutionQueue = (*RedisExecutionQueue)(nil) // Ensure interface is implemented at compile time
