package dao_adapters_test

import (
	dao_adapters "brokerx/matching-service/adapters/dao"
	"brokerx/matching-service/mocks"
	"brokerx/matching-service/models"
	"context"
	"testing"
)

var ctx = context.Background()
var redisClientMock, s = mocks.GetRedisClientMock()
var orderBook = dao_adapters.RedisOrderBook{Rdb: redisClientMock}

var order1LimitBuy = models.Order{
	ID:                1,
	UserID:            1,
	Symbol:            "AAPL",
	Type:              "limit",
	Action:            "buy",
	Quantity:          13,
	RemainingQuantity: 13,
	UnitPrice:         125.0,
	Timing:            "day",
	Status:            "open",
}
var order2LimitSell = models.Order{
	ID:                2,
	UserID:            2,
	Symbol:            "AAPL",
	Type:              "limit",
	Action:            "sell",
	Quantity:          13,
	RemainingQuantity: 13,
	UnitPrice:         127.0,
	Timing:            "day",
	Status:            "open",
}
var order3MarketBuy = models.Order{
	ID:                3,
	UserID:            2,
	Symbol:            "AAPL",
	Type:              "market",
	Action:            "buy",
	Quantity:          13,
	RemainingQuantity: 13,
	Timing:            "day",
	Status:            "open",
}
var order4MarketSell = models.Order{
	ID:                4,
	UserID:            2,
	Symbol:            "AAPL",
	Type:              "market",
	Action:            "sell",
	Quantity:          13,
	RemainingQuantity: 13,
	Timing:            "day",
	Status:            "open",
}

var baseOrders = []*models.Order{&order1LimitBuy, &order2LimitSell, &order3MarketBuy, &order4MarketSell}

var insertOrder = models.Order{
	ID:                5,
	UserID:            2,
	Symbol:            "AAPL",
	Type:              "market",
	Action:            "sell",
	Quantity:          13,
	RemainingQuantity: 13,
	Timing:            "day",
	Status:            "open",
}

func setupOrder(wantErr bool) {
	s.Close()
	if !wantErr {
		redisClientMock, s = mocks.GetRedisClientMock()
		orderBook = dao_adapters.RedisOrderBook{Rdb: redisClientMock}
		for _, o := range baseOrders {
			_ = orderBook.Insert(ctx, o)
		}
	}
}

func orderGoneFromBook(orderId int) bool {
	order, _ := orderBook.GetById(ctx, orderId)
	return order.ID == 0
}

// func matchLimitCandidatesValid(candidates []*models.Order, wantLength int, wantTypes []string, wantAction string, wantUnitPrice float64) bool {
// 	if len(candidates) != wantLength {
// 		log.Error("didnt receive expected amount of candidates")
// 		return false
// 	}

// 	for _, c := range candidates {
// 		if !slices.Contains(wantTypes, c.Type) {
// 			log.Error("candidate i wrong type")
// 			return false
// 		}
// 		if c.Action != wantAction {
// 			log.Error("candidate is wrong action")
// 			return false
// 		}
// 		if c.Action == "sell" && c.Type != "market" && len(wantTypes) == 2 && c.UnitPrice > wantUnitPrice {
// 			log.Error("sell candidate is wrong unit price")
// 			return false
// 		}
// 		if c.Action == "buy" && c.Type != "market" && len(wantTypes) == 2 && c.UnitPrice < wantUnitPrice {
// 			log.Error("buy candidate is wrong unit price")
// 			return false
// 		}
// 	}

// 	return true
// }

// func matchMarketCandidatesValid(candidates []*models.Order, wantLength int, wantType, wantAction string) bool {
// 	if wantLength != len(candidates) {
// 		return false
// 	}
// 	for _, c := range candidates {
// 		if c.Type != wantType || c.Action != wantAction {
// 			return false
// 		}
// 	}
// 	return true
// }

// func checkOrderQueueLength(expected int64) {
// 	length, _ := redisClientMock.LLen(context.Background(), dao_adapters.ORDER_PERSISTANCE_QUEUE).Result()
// 	if length != expected {
// 		panic(fmt.Sprintf("Order queue length() = %v, want %v", length, expected))
// 	}
// }

func TestRedisOrderBook_GetById(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		orderId int
		want    models.Order
		wantErr bool
	}{
		{
			name:    "returns order with specified id and removes it from book",
			orderId: 1,
			want:    order1LimitBuy,
			wantErr: false,
		},
		{
			name:    "returns no order when no order with specified id",
			orderId: 10,
			want:    models.Order{},
			wantErr: false,
		},
		{
			name:    "returns error when redis is down",
			orderId: 0,
			want:    models.Order{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupOrder(tt.wantErr)

			got, gotErr := orderBook.GetById(ctx, tt.orderId)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetById() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetById() succeeded unexpectedly")
			}
			if got.ID != tt.want.ID || !orderGoneFromBook(tt.orderId) {
				t.Errorf("GetById() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestRedisOrderBook_FindMatchesLimit(t *testing.T) {
// 	tests := []struct {
// 		name string // description of this test case
// 		// Named input parameters for target function.
// 		symbol        string
// 		action        string
// 		unitPrice     float64
// 		batchSize     int
// 		wantLength    int
// 		wantTypes     []string
// 		wantAction    string
// 		wantUnitPrice float64
// 		wantErr       bool
// 	}{
// 		{
// 			name:          "incoming limit buy order returns match candidates",
// 			symbol:        "AAPL",
// 			action:        "buy",
// 			unitPrice:     128.0,
// 			batchSize:     10,
// 			wantLength:    2,
// 			wantTypes:     []string{"market", "limit"},
// 			wantAction:    "sell",
// 			wantUnitPrice: 128.0,
// 			wantErr:       false,
// 		},
// 		{
// 			name:          "incoming limit sell order returns match candidates",
// 			symbol:        "AAPL",
// 			action:        "sell",
// 			unitPrice:     120.0,
// 			batchSize:     10,
// 			wantLength:    2,
// 			wantTypes:     []string{"market", "limit"},
// 			wantAction:    "buy",
// 			wantUnitPrice: 120.0,
// 			wantErr:       false,
// 		},
// 		{
// 			name:    "returns error when redis is down",
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			setupOrder(tt.wantErr)

// 			got, gotErr := orderBook.FindMatchesLimit(ctx, tt.symbol, tt.action, tt.unitPrice, tt.batchSize)

// 			if gotErr != nil {
// 				if !tt.wantErr {
// 					t.Errorf("FindMatchesLimit() failed: %v", gotErr)
// 				}
// 				return
// 			}
// 			if tt.wantErr {
// 				t.Fatal("FindMatchesLimit() succeeded unexpectedly")
// 			}
// 			if !matchLimitCandidatesValid(got, tt.wantLength, tt.wantTypes, tt.wantAction, tt.wantUnitPrice) {
// 				t.Errorf("FindMatchesLimit() = %v, want not that", got)
// 			}
// 		})
// 	}
// }

// func TestRedisOrderBook_FindMatchesMarket(t *testing.T) {
// 	tests := []struct {
// 		name string // description of this test case
// 		// Named input parameters for target function.
// 		symbol     string
// 		action     string
// 		unitPrice  float64
// 		batchSize  int
// 		wantLength int
// 		wantType   string
// 		wantAction string
// 		wantErr    bool
// 	}{

// 		{
// 			name:       "incoming market buy order returns match candidates",
// 			symbol:     "AAPL",
// 			action:     "buy",
// 			unitPrice:  128.0,
// 			batchSize:  10,
// 			wantLength: 1,
// 			wantType:   "limit",
// 			wantAction: "sell",
// 			wantErr:    false,
// 		},
// 		{
// 			name:       "incoming market sell order returns match candidates",
// 			symbol:     "AAPL",
// 			action:     "sell",
// 			unitPrice:  119.0,
// 			batchSize:  10,
// 			wantLength: 1,
// 			wantType:   "limit",
// 			wantAction: "buy",
// 			wantErr:    false,
// 		},
// 		{
// 			name:    "returns error when redis is down",
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			setupOrder(tt.wantErr)

// 			got, gotErr := orderBook.FindMatchesMarket(ctx, tt.symbol, tt.action, tt.batchSize)

// 			if gotErr != nil {
// 				if !tt.wantErr {
// 					t.Errorf("FindMatchesLimit() failed: %v", gotErr)
// 				}
// 				return
// 			}
// 			if tt.wantErr {
// 				t.Fatal("FindMatchesLimit() succeeded unexpectedly")
// 			}
// 			if !matchMarketCandidatesValid(got, tt.wantLength, tt.wantType, tt.wantAction) {
// 				t.Errorf("FindMatchesLimit() = %v, want not that", got)
// 			}
// 		})
// 	}
// }

func TestRedisOrderBook_Insert(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		order   *models.Order
		wantErr bool
	}{
		{
			name:  "order is inserted into order book",
			order: &insertOrder,
		},
		{
			name:    "returns error when redis is down",
			order:   &models.Order{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupOrder(tt.wantErr)

			gotErr := orderBook.Insert(ctx, tt.order)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Insert() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Insert() succeeded unexpectedly")
			}
			o, err := orderBook.GetById(ctx, tt.order.ID)
			if err != nil || o.ID != tt.order.ID {
				t.Errorf("Insert() didnt insert order into order book : %v", err)
			}
		})
	}
}

// func TestRedisOrderBook_Return(t *testing.T) {
// 	tests := []struct {
// 		name string // description of this test case
// 		// Named input parameters for target function.
// 		orders  []*models.Order
// 		wantErr bool
// 	}{
// 		{
// 			name: "orders are all returned to order book",
// 			orders: []*models.Order{
// 				&order1LimitBuy,
// 				&order2LimitSell,
// 			},
// 		},
// 		{
// 			name:    "returns error when redis is down",
// 			orders:  []*models.Order{&order4MarketSell},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			setupOrder(tt.wantErr)

// 			gotErr := orderBook.Return(ctx, tt.orders)

// 			if gotErr != nil {
// 				if !tt.wantErr {
// 					t.Errorf("Return() failed: %v", gotErr)
// 				}
// 				return
// 			}
// 			if tt.wantErr {
// 				t.Fatal("Return() succeeded unexpectedly")
// 			}
// 			for _, order := range tt.orders {
// 				o, err := orderBook.GetById(ctx, order.ID)
// 				if err != nil || o.ID != order.ID {
// 					t.Errorf("Insert() didnt return orders into order book : %v", err)
// 				}
// 			}
// 		})
// 	}
// }
