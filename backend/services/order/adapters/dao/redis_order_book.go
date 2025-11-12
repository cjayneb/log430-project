package dao_adapters

import (
	"brokerx/order-service/common"
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const ORDER_PERSISTANCE_QUEUE = "orderpersistancequeue"

type RedisOrderBook struct {
	Rdb *redis.Client
}

func (book *RedisOrderBook) DequeueOrders(ctx context.Context, batchSize int) ([]*models.Order, error) {
	log := common.FromContext(ctx)

	var orders []*models.Order
	for i := 0; i < batchSize; i++ {
		val, err := book.Rdb.RPop(ctx, ORDER_PERSISTANCE_QUEUE).Result()
		if err == redis.Nil {
			break // queue is empty
		}
		if err != nil {
			log.Error("error dequeuing orders", "error", err)
			return nil, err
		}
		var order models.Order
		if err := json.Unmarshal([]byte(val), &order); err != nil {
			log.Error("error unmarshaling dequeued order", "error", err)
			continue
		}
		orders = append(orders, &order)
	}
	return orders, nil
}

func (book *RedisOrderBook) FetchByIDs(ctx context.Context, ids []string) ([]*models.Order, error) {
	return book.fetchOrders(ctx, ids, false)
}

func (book *RedisOrderBook) fetchOrders(ctx context.Context, ids []string, fetchAndDelete bool) ([]*models.Order, error) {
	log := common.FromContext(ctx)

	if len(ids) == 0 {
		return []*models.Order{}, nil
	}

	pipe := book.Rdb.Pipeline()
	cmds := make([]*redis.StringCmd, len(ids))
	for i, id := range ids {
		if fetchAndDelete {
			cmds[i] = pipe.GetDel(ctx, keyOrderStr(id))
		} else {
			cmds[i] = pipe.Get(ctx, keyOrderStr(id))
		}
	}
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		log.Error("error executing command to fetch orders", "error", err)
		return nil, err
	}

	results := make([]*models.Order, 0, len(ids))
	for _, cmd := range cmds {
		val, err := cmd.Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			log.Warn("error executing command", "error", err)
			continue
		}
		var o models.Order
		if err := json.Unmarshal([]byte(val), &o); err == nil {
			results = append(results, &o)
		} else {
			log.Warn("error unmarshaling JSON into order", "error", err)
		}
	}
	return results, nil
}

func keyOrderStr(id string) string {
	return fmt.Sprintf("order:%s", id)
}

var _ ports.OrderBook = (*RedisOrderBook)(nil) // Ensure interface is implemented at compile time
