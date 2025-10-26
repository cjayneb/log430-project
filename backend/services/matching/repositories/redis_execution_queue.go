package repositories

import (
	"brokerx/matching-service/models"
	"brokerx/matching-service/ports"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

const EXECUTION_RECORDS_PERSISTANCE_QUEUE = "execrecordspersistancequeue"

type RedisExecutionQueue struct {
	Rdb *redis.Client
}

func (queue *RedisExecutionQueue) EnqueueExecutionRecords(records []*models.ExecutionRecord) error {
	for _, record := range records {
		data, err := json.Marshal(record)
		if err != nil {
			log.Errorf("error marshaling execution record %v for enqueue: %v", record, err)
			continue
		}
		if err := queue.Rdb.LPush(ctx, EXECUTION_RECORDS_PERSISTANCE_QUEUE, data).Err(); err != nil {
			log.Errorf("error enqueueing execution record %v: %v", record, err)
		}
	}
	return nil
}

func (queue *RedisExecutionQueue) DequeueExecutionRecords(batchSize int) ([]*models.ExecutionRecord, error) {
	var records []*models.ExecutionRecord
	for i := 0; i < batchSize; i++ {
		val, err := queue.Rdb.RPop(ctx, EXECUTION_RECORDS_PERSISTANCE_QUEUE).Result()
		if err == redis.Nil {
			break // queue is empty
		}
		if err != nil {
			return nil, err
		}
		var record models.ExecutionRecord
		if err := json.Unmarshal([]byte(val), &record); err != nil {
			log.Errorf("error unmarshaling dequeued execution record: %v", err)
			continue
		}
		records = append(records, &record)
	}
	return records, nil
}

var _ ports.ExecutionQueue = (*RedisExecutionQueue)(nil) // Ensure interface is implemented at compile time
