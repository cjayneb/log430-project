package dao_adapters

import (
	"brokerx/matching-service/models"
	"brokerx/matching-service/ports"
	"brokerx/matching-service/util"
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

const EXECUTION_RECORDS_PERSISTANCE_QUEUE = "execrecordspersistancequeue"

type RedisExecutionQueue struct {
	Rdb *redis.Client
}

func (queue *RedisExecutionQueue) EnqueueExecutionRecords(ctx context.Context, records []*models.ExecutionRecord) error {
	log := util.FromContext(ctx)

	for _, record := range records {
		data, err := json.Marshal(record)
		if err != nil {
			log.Error("error marshaling execution record for enqueue", "record", record, "error", err)
			continue
		}
		if err := queue.Rdb.LPush(ctx, EXECUTION_RECORDS_PERSISTANCE_QUEUE, data).Err(); err != nil {
			log.Error("error enqueueing execution record", "record", record, "error", err)
			return err
		}
	}
	return nil
}

var _ ports.ExecutionQueue = (*RedisExecutionQueue)(nil) // Ensure interface is implemented at compile time
