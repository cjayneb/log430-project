package dao_adapters

import (
	"brokerx/matching-service/models"
	"brokerx/matching-service/ports"
	"brokerx/matching-service/util"
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

const ORDER_PERSISTANCE_QUEUE = "orderpersistancequeue"
const PRICE_MULT = float64(1e12)

var insertScript = redis.NewScript(`
	-- KEYS:
	-- 1 = orderKey
	-- 2 = orderBookKey

	-- ARGV:
	-- 1 = rawOrderStr
	-- 2 = addToSortedSet
	-- 3 = sortedSetScore (unit price or timestamp)
	-- 4 = orderId

	redis.call("SET", KEYS[1], ARGV[1])
	if ARGV[2] == "1" then
		redis.call("ZADD", KEYS[2], ARGV[3], ARGV[4])
	end
	return 1
`)

var getAndDeleteOrderScript = redis.NewScript(`
	local raw = redis.call("GET", KEYS[1])
	if not raw then
        return nil
    end
	local order = cjson.decode(raw)

	if not order.sideKey or type(order.sideKey) ~= "string" then
		return {err="Missing or invalid sideKey"}
	end

	local id = order.order_id
	if type(id) ~= "number" then
		return {err="Invalid order id"}
	end

	redis.call("DEL", KEYS[1])
	redis.call("ZREM", order.sideKey, tostring(id))

	return raw
`)

var getMatchesAndPopScript = redis.NewScript(`
	-- KEYS:
	-- 1 = limitKey
	-- 2 = marketKey

	-- ARGV:
	-- 1 = action ("buy" or "sell")
	-- 2 = unitPrice
	-- 3 = batchSize
	-- 4 = orderType ("limit" or "market")
	-- 5 = userId

	local action = ARGV[1]
	local price = tonumber(ARGV[2])
	local batchSize = tonumber(ARGV[3])
	local orderType = ARGV[4]
	local userId = tostring(ARGV[5])

	local ids = {}

	if action == "buy" then
		-- BUY order -> consume lowest sells
		ids = redis.call("ZRANGE", KEYS[1], 0, batchSize - 1)
	else
		-- SELL order -> consume highest buys
		ids = redis.call("ZRANGE", KEYS[1], -batchSize, -1, "REV")
	end

	local results = {}
	for _, id in ipairs(ids) do
		-- compute hash key
		local hashKey = "order:" .. id
		local raw = redis.call("GET", hashKey)

		if raw then
			local order = cjson.decode(raw)

			if (userId ~= tostring(order.user_id)) and ((orderType == "market") or (action == "buy" and order.unit_price <= price) or (action == "sell" and order.unit_price >= price)) then
				-- delete the order hash
				redis.call("DEL", hashKey)

				-- delete from sorted sets
				if order.sideKey then
					redis.call("ZREM", order.sideKey, id)
				end
				redis.call("ZREM", KEYS[1], id)
				redis.call("ZREM", KEYS[2], id)

				table.insert(results, raw)
			end
		end
	end

	-- Add market orders if not enough limit candidates
	if orderType == "limit" and #results < batchSize then
		batchSize = batchSize - #results
		local marketIds = redis.call("ZRANGE", KEYS[2], 0, batchSize)

		for _, id in ipairs(marketIds) do
			-- compute hash key
			local hashKey = "order:" .. id
			local raw = redis.call("GET", hashKey)

			if raw then
				local order = cjson.decode(raw)

				if (userId ~= tostring(order.user_id)) then
					-- delete the order hash
					redis.call("DEL", hashKey)

					-- delete from sorted sets
					if order.sideKey then
						redis.call("ZREM", order.sideKey, id)
					end
					redis.call("ZREM", KEYS[1], id)
					redis.call("ZREM", KEYS[2], id)

					table.insert(results, raw)
				end
			end
		end
	end

	return results
`)

type RedisOrderBook struct {
	Rdb *redis.Client
}

func (book *RedisOrderBook) FindMatches(ctx context.Context, order *models.Order, batchSize int) ([]*models.Order, error) {
	log := util.FromContext(ctx)

	opposite := "sell"
	if order.Action == "sell" {
		opposite = "buy"
	}

	limitKey := keyBook(order.Symbol, opposite, "limit")
	marketKey := keyBook(order.Symbol, opposite, "market")

	res, err := getMatchesAndPopScript.Run(ctx, book.Rdb,
		[]string{limitKey, marketKey},
		order.Action,
		order.UnitPrice,
		batchSize,
		order.Type,
		order.UserID,
	).Result()
	if err != nil {
		log.Error("error when executing atomic fetch/pop candidates", "error", err)
		return nil, err
	}

	rawOrders := res.([]interface{})

	orders := make([]*models.Order, 0, len(rawOrders))
	for _, raw := range rawOrders {
		var o models.Order
		if err := json.Unmarshal([]byte(raw.(string)), &o); err == nil {
			orders = append(orders, &o)
		} else {
			log.Error("error unmarshaling popped candidate order", "rawOrder", raw, "error", err)
		}
	}

	return orders, nil
}

func (book *RedisOrderBook) GetById(ctx context.Context, orderId int) (models.Order, error) {
	log := util.FromContext(ctx)

	rawOrder, err := getAndDeleteOrderScript.Run(ctx, book.Rdb, []string{keyOrder(orderId)}).Result()
	if err == redis.Nil {
		return models.Order{}, nil
	}
	if err != nil {
		log.Error("error when executing atomic get/remove", "orderId", orderId, "error", err)
		return models.Order{}, err
	}

	var order models.Order
	if err := json.Unmarshal([]byte(rawOrder.(string)), &order); err != nil {
		log.Error("error when unmarshaling order from order hashes", "orderId", orderId, "error", err)
		return models.Order{}, err
	}

	return order, nil
}

func (book *RedisOrderBook) Insert(ctx context.Context, order *models.Order) error {
	sideKey := keyBook(order.Symbol, order.Action, order.Type)
	order.SideKey = sideKey
	data, err := json.Marshal(order)
	if err != nil {
		log.Error("error when marshaling order into JSON", "error", err)
		return err
	}

	key := keyOrder(order.ID)

	addToSortedSet := "1"
	if order.Timing == "ioc" {
		// IOC orders should not be added to order book sorted set
		addToSortedSet = "0"
	}

	return insertScript.Run(ctx, book.Rdb,
		[]string{key, sideKey},
		string(data),
		addToSortedSet,
		computeScore(order),
		order.ID,
	).Err()
}

func computeScore(order *models.Order) float64 {
	ts := float64(order.CreatedAt.Time.UnixNano())

	if order.Type == "market" {
		return ts
	}

	priceMajor := order.UnitPrice * PRICE_MULT

	if order.Action == "buy" {
		return priceMajor - ts
	} else {
		return priceMajor + ts
	}
}

func (book *RedisOrderBook) Return(ctx context.Context, orders []*models.Order) {
	log := util.FromContext(ctx)

	// TODO: Send orders that could not be returned to a recovery queue that will return them safely
	for _, order := range orders {
		if err := book.Insert(ctx, order); err != nil {
			log.Error("error when returning order to order book", "orderId", order.ID, "error", err)
			continue
		}
	}
}

func keyOrder(id int) string {
	return fmt.Sprintf("order:%d", id)
}

func keyBook(symbol, side, orderType string) string {
	return fmt.Sprintf("orderbook:%s:%s:%s", symbol, side, orderType)
}

var _ ports.OrderBook = (*RedisOrderBook)(nil) // Ensure interface is implemented at compile time
