package repositories

import (
	"brokerx/matching-service/models"
	"brokerx/matching-service/ports"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

const ORDER_PERSISTANCE_QUEUE = "orderpersistancequeue"

type RedisOrderBook struct {
	Rdb *redis.Client
}

var ctx = context.Background()

func (book *RedisOrderBook) GetById(orderId int) (models.Order, error) {
	val, err := book.Rdb.GetDel(ctx, keyOrder(orderId)).Result()
	if err == redis.Nil {
		return models.Order{}, nil
	}
	if err != nil {
		return models.Order{}, err
	}

	var order models.Order
	if err := json.Unmarshal([]byte(val), &order); err != nil {
		return models.Order{}, err
	}

	// remove from order book sorted set so only one worker processes it
	book.Rdb.ZRem(ctx, keyBook(order.Symbol, order.Action, order.Type), orderId)

	return order, nil
}

func (book *RedisOrderBook) FindMatchesLimit(symbol string, orderType string, action string, unitPrice float64, batchSize int) ([]*models.Order, error) {
	opposite := "sell"
	if action == "sell" {
		opposite = "buy"
	}

	limitKey := keyBook(symbol, opposite, "limit")
	marketKey := keyBook(symbol, opposite, "market")

	var ids []string
	var err error

	if action == "buy" {
		// Find sells <= buy price
		ids, err = book.Rdb.ZRangeByScore(ctx, limitKey, &redis.ZRangeBy{
			Min:   "-inf",
			Max:   fmt.Sprintf("%f", unitPrice),
			Count: int64(batchSize),
		}).Result()
	} else {
		// Find buys >= sell price
		ids, err = book.Rdb.ZRevRangeByScore(ctx, limitKey, &redis.ZRangeBy{
			Max:   "+inf",
			Min:   fmt.Sprintf("%f", unitPrice),
			Count: int64(batchSize),
		}).Result()
	}
	if err != nil {
		return nil, err
	}

	// Also include market orders on opposite side (they take any price)
	marketIDs, err := book.Rdb.ZRange(ctx, marketKey, 0, int64(batchSize-1)).Result()
	if err == nil {
		ids = append(ids, marketIDs...)
	}

	for _, id := range ids {
		book.Rdb.ZRem(ctx, limitKey, id)
		book.Rdb.ZRem(ctx, marketKey, id)
	}

	return book.fetchOrders(ids, true)
}

func (book *RedisOrderBook) FindMatchesMarket(symbol string, orderType string, action string, batchSize int) ([]*models.Order, error) {
	opposite := "sell"
	if action == "sell" {
		opposite = "buy"
	}

	limitKey := keyBook(symbol, opposite, "limit")

	var popped []redis.Z
	var err error

	if action == "buy" {
		// Buyer wants lowest sells (ZPOPMIN)
		popped, err = book.Rdb.ZPopMin(ctx, limitKey, int64(batchSize)).Result()
	} else {
		// Seller wants highest buys (ZPOPMAX)
		popped, err = book.Rdb.ZPopMax(ctx, limitKey, int64(batchSize)).Result()
	}
	if err != nil && err != redis.Nil {
		return nil, err
	}

	ids := make([]string, 0, len(popped))
	for _, z := range popped {
		ids = append(ids, fmt.Sprintf("%v", z.Member))
	}
	return book.fetchOrders(ids, true)
}

func (book *RedisOrderBook) Insert(order *models.Order) error {
	key := keyOrder(order.ID)
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}
	if err := book.Rdb.Set(ctx, key, data, 0).Err(); err != nil {
		return err
	}

	// market orders: score = timestamp (for FIFO)
	// limit orders: score = price
	score := order.UnitPrice
	if order.Type == "market" {
		score = float64(time.Now().UnixNano())
	}

	if order.Timing == "ioc" {
		// IOC orders should not be added to order book sorted set
		return nil
	}

	sideKey := keyBook(order.Symbol, order.Action, order.Type)
	return book.Rdb.ZAdd(ctx, sideKey, redis.Z{Score: score, Member: order.ID}).Err()
}

func (book *RedisOrderBook) Return(orders []*models.Order) error {
	for _, order := range orders {
		if err := book.Insert(order); err != nil {
			return err
		}
		if err := book.MarkDirty(order.ID); err != nil {
			return err
		}
	}
	return nil
}

func (book *RedisOrderBook) MarkDirty(orderID int) error {
	return book.Rdb.LPush(ctx, "orders:dirty", orderID).Err()
}

func (book *RedisOrderBook) EnqueueOrders(orders []*models.Order) error {
	for _, order := range orders {
		data, err := json.Marshal(order)
		if err != nil {
			log.Errorf("error marshaling order %v for enqueue: %v", order, err)
			continue
		}
		if err := book.Rdb.LPush(ctx, ORDER_PERSISTANCE_QUEUE, data).Err(); err != nil {
			log.Errorf("error enqueueing order %v: %v", order, err)
		}
	}
	return nil
}

func (book *RedisOrderBook) DequeueOrders(batchSize int) ([]*models.Order, error) {
	var orders []*models.Order
	for i := 0; i < batchSize; i++ {
		val, err := book.Rdb.RPop(ctx, ORDER_PERSISTANCE_QUEUE).Result()
		if err == redis.Nil {
			break // queue is empty
		}
		if err != nil {
			return nil, err
		}
		var order models.Order
		if err := json.Unmarshal([]byte(val), &order); err != nil {
			log.Errorf("error unmarshaling dequeued order: %v", err)
			continue
		}
		orders = append(orders, &order)
	}
	return orders, nil
}

func (book *RedisOrderBook) FetchByIDs(ids []string) ([]*models.Order, error) {
	return book.fetchOrders(ids, false)
}

func (book *RedisOrderBook) LogBook() {
	orders, _ := book.fetchAll()
	log.Info()
	log.Info("Contents of Redis Set")
	for _, order := range orders {
		log.Infof("Redis order: %v", order)
	}
	log.Info("End of Redis Set---------------")
	log.Info()
}

func (book *RedisOrderBook) fetchAll() ([]*models.Order, error) {
	var orders []*models.Order
	iter := book.Rdb.Scan(ctx, 0, "order:*", 0).Iterator()
	for iter.Next(ctx) {
		val, err := book.Rdb.Get(ctx, iter.Val()).Result()
		if err != nil {
			continue
		}
		var o models.Order
		if err := json.Unmarshal([]byte(val), &o); err == nil {
			orders = append(orders, &o)
		}
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func (book *RedisOrderBook) fetchOrders(ids []string, fetchAndDelete bool) ([]*models.Order, error) {
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
		return nil, err
	}

	results := make([]*models.Order, 0, len(ids))
	for _, cmd := range cmds {
		val, err := cmd.Result()
		if err != nil {
			continue
		}
		var o models.Order
		if err := json.Unmarshal([]byte(val), &o); err == nil {
			results = append(results, &o)
		}
	}
	return results, nil
}

func keyOrder(id int) string {
	return fmt.Sprintf("order:%d", id)
}

func keyOrderStr(id string) string {
	return fmt.Sprintf("order:%s", id)
}

func keyBook(symbol, side, orderType string) string {
	return fmt.Sprintf("orderbook:%s:%s:%s", symbol, side, orderType)
}

var _ ports.OrderBook = (*RedisOrderBook)(nil) // Ensure interface is implemented at compile time
