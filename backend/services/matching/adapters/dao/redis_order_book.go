package dao_adapters

import (
	"brokerx/matching-service/models"
	"brokerx/matching-service/ports"
	"brokerx/matching-service/util"
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

func (book *RedisOrderBook) GetById(ctx context.Context, orderId int) (models.Order, error) {
	log := util.FromContext(ctx)

	val, err := book.Rdb.GetDel(ctx, keyOrder(orderId)).Result()
	if err == redis.Nil {
		return models.Order{}, nil
	}
	if err != nil {
		log.Error("error when fetching order from order hashes", "orderId", orderId, "error", err)
		return models.Order{}, err
	}

	var order models.Order
	if err := json.Unmarshal([]byte(val), &order); err != nil {
		log.Error("error when unmarshaling order from order hashes", "orderId", orderId, "error", err)
		return models.Order{}, err
	}

	// remove from order book sorted set so only one worker processes it
	if err := book.Rdb.ZRem(ctx, keyBook(order.Symbol, order.Action, order.Type), orderId).Err(); err != nil {
		log.Error("error when removing order from order book sorted set", "orderId", orderId, "error", err)
		return models.Order{}, err
	}

	return order, nil
}

func (book *RedisOrderBook) FindMatchesLimit(ctx context.Context, symbol string, action string, unitPrice float64, batchSize int) ([]*models.Order, error) {
	log := util.FromContext(ctx)

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
		log.Error("error when fetching matches from order book", "error", err)
		return nil, err
	}

	// Also include market orders on opposite side (they take any price)
	marketIDs, err := book.Rdb.ZRange(ctx, marketKey, 0, int64(batchSize-1)).Result()
	if err != nil {
		log.Warn("could not fetch market match candidates from order book", "error", err)
	} else {
		ids = append(ids, marketIDs...)
	}

	for _, id := range ids {
		book.Rdb.ZRem(ctx, limitKey, id)
		book.Rdb.ZRem(ctx, marketKey, id)
	}

	return book.fetchOrders(ctx, ids, true)
}

func (book *RedisOrderBook) FindMatchesMarket(ctx context.Context, symbol string, action string, batchSize int) ([]*models.Order, error) {
	log := util.FromContext(ctx)

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
		log.Error("error when fetching matches from order book", "error", err)
		return nil, err
	}

	ids := make([]string, 0, len(popped))
	for _, z := range popped {
		ids = append(ids, fmt.Sprintf("%v", z.Member))
	}
	return book.fetchOrders(ctx, ids, true)
}

func (book *RedisOrderBook) Insert(ctx context.Context, order *models.Order) error {
	key := keyOrder(order.ID)
	data, err := json.Marshal(order)
	if err != nil {
		log.Error("error when marshaling order into JSON", "error", err)
		return err
	}
	if err := book.Rdb.Set(ctx, key, data, 0).Err(); err != nil {
		log.Error("error when inserting order into order hashes", "error", err)
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

func (book *RedisOrderBook) Return(ctx context.Context, orders []*models.Order) error {
	log := util.FromContext(ctx)

	for _, order := range orders {
		if err := book.Insert(ctx, order); err != nil {
			log.Error("error when returning order to order book", "orderId", order.ID, "error", err)
			return err
		}
		if err := book.MarkDirty(ctx, order.ID); err != nil {
			log.Error("error when marking order as dirty", "orderId", order.ID, "error", err)
			return err
		}
	}
	return nil
}

func (book *RedisOrderBook) MarkDirty(ctx context.Context, orderID int) error {
	return book.Rdb.LPush(ctx, "orders:dirty", orderID).Err()
}

func (book *RedisOrderBook) EnqueueOrders(ctx context.Context, orders []*models.Order) error {
	log := util.FromContext(ctx)

	for _, order := range orders {
		data, err := json.Marshal(order)
		if err != nil {
			log.Error("error marshaling order for enqueue", "orderId", order.ID, "error", err)
			continue
		}
		if err := book.Rdb.LPush(ctx, ORDER_PERSISTANCE_QUEUE, data).Err(); err != nil {
			log.Error("error enqueueing order", "orderId", order.ID, "error", err)
			return err
		}
	}
	return nil
}

func (book *RedisOrderBook) fetchOrders(ctx context.Context, ids []string, fetchAndDelete bool) ([]*models.Order, error) {
	log := util.FromContext(ctx)

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
		log.Error("error when fetching orders from order hashes", "error", err)
		return nil, err
	}

	results := make([]*models.Order, 0, len(ids))
	for _, cmd := range cmds {
		val, err := cmd.Result()
		if err != nil {
			log.Warn("error when reading command result", "error", err)
			continue
		}
		var o models.Order
		err = json.Unmarshal([]byte(val), &o)
		if err != nil {
			log.Warn("error when unmarshaling fetched order", "error", err)
		} else {
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
