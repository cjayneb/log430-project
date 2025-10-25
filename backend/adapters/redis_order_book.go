package adapters

import (
	"brokerx/models"
	"brokerx/ports"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisOrderBook struct {
	Rdb *redis.Client
}

var ctx = context.Background()

func keyOrder(id int) string {
	return fmt.Sprintf("order:%d", id)
}

func keyOrderStr(id string) string {
	return fmt.Sprintf("order:%s", id)
}

func keyBook(symbol, side, orderType string) string {
	return fmt.Sprintf("orderbook:%s:%s:%s", symbol, side, orderType)
}

func (book *RedisOrderBook) MarkDirty(orderID int) error {
	return book.Rdb.LPush(ctx, "orders:dirty", orderID).Err()
}

func (book *RedisOrderBook) GetById(orderId int) (models.Order, error) {
	val, err := book.Rdb.Get(ctx, keyOrder(orderId)).Result()
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

	// remove from Redis so only one worker processes it
	book.Rdb.Del(ctx, keyOrder(orderId))
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

	return book.fetchOrders(ids)
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
	return book.fetchOrders(ids)
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

	sideKey := keyBook(order.Symbol, order.Action, order.Type)
	if err := book.Rdb.ZAdd(ctx, sideKey, redis.Z{Score: score, Member: order.ID}).Err(); err != nil {
		return err
	}

	// Mark this order as dirty for MySQL sync
	return book.MarkDirty(order.ID)
}

func (book *RedisOrderBook) Return(orders []*models.Order) error {
	for _, order := range orders {
		if err := book.Insert(order); err != nil {
			return err
		}
	}
	return nil
}

func (book *RedisOrderBook) FetchByIDs(ids []string) ([]*models.Order, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	pipe := book.Rdb.Pipeline()
	cmds := make([]*redis.StringCmd, len(ids))
	for i, id := range ids {
		cmds[i] = pipe.Get(ctx, keyOrderStr(id))
	}
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	results := []*models.Order{}
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

func (book *RedisOrderBook) fetchOrders(ids []string) ([]*models.Order, error) {
	pipe := book.Rdb.Pipeline()
	cmds := make([]*redis.StringCmd, len(ids))
	for i, id := range ids {
		cmds[i] = pipe.GetDel(ctx, keyOrderStr(id))
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

var _ ports.OrderBook = (*RedisOrderBook)(nil) // Ensure interface is implemented at compile time
