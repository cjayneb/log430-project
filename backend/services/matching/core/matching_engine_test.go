package core_test

import (
	"brokerx/matching-service/core"
	"brokerx/matching-service/mocks"
	"brokerx/matching-service/models"
	"context"
	"errors"
	"os"
	"testing"
)

var ctx = context.Background()
var engine core.MatchingEngineImpl

var order1LimitBuy = models.Order{
	ID:                1,
	UserID:            1,
	Symbol:            "AAPL",
	Type:              "limit",
	Action:            "buy",
	Quantity:          13,
	RemainingQuantity: 10,
	UnitPrice:         125.0,
	Timing:            "day",
	Status:            "partially_filled",
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
	UserID:            3,
	Symbol:            "AAPL",
	Type:              "market",
	Action:            "buy",
	Quantity:          13,
	RemainingQuantity: 13,
	Timing:            "day",
	Status:            "open",
}
var order4MarketBuy = models.Order{
	ID:                4,
	UserID:            3,
	Symbol:            "AAPL",
	Type:              "market",
	Action:            "buy",
	Quantity:          13,
	RemainingQuantity: 0,
	Timing:            "day",
	Status:            "filled",
}

func TestMain(m *testing.M) {
	engine = core.MatchingEngineImpl{}
	engine.StartMatchingWorkers(0)
	code := m.Run()
	os.Exit(code)
}

func TestMatchingEngineImpl_QueueOrder(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		mockOrderBook mocks.MockOrderBook
		mockExecQueue mocks.MockExecQueue
		order         *models.Order
		wantLength    int
		wantErr       bool
	}{
		{
			name:          "order 1 queued sucessfully",
			mockOrderBook: mocks.MockOrderBook{},
			mockExecQueue: mocks.MockExecQueue{},
			order:         &models.Order{},
			wantLength:    1,
			wantErr:       false,
		},
		{
			name:          "order 2 queued sucessfully",
			mockOrderBook: mocks.MockOrderBook{},
			mockExecQueue: mocks.MockExecQueue{},
			order:         &models.Order{},
			wantLength:    2,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine.OrderBook = &tt.mockOrderBook
			engine.ExecutionQueue = &tt.mockExecQueue

			gotErr := engine.QueueOrder(ctx, tt.order)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("QueueOrder() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("QueueOrder() succeeded unexpectedly")
			}
			if len(core.OrderQueue) != tt.wantLength {
				t.Errorf("order was not queued")
			}
		})
	}
}

func TestMatchingEngineImpl_SubmitOrder(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		orderId         int
		matchCandidates []*models.Order
		mockOrderBook   mocks.MockOrderBook
		mockExecQueue   mocks.MockExecQueue
		wantErr         bool
	}{
		{
			name:          "order stays open",
			orderId:       1,
			mockOrderBook: mocks.MockOrderBook{Order: order1LimitBuy, Orders: []*models.Order{}},
		},
		{
			name:          "error is returned when order book fails",
			orderId:       1,
			mockOrderBook: mocks.MockOrderBook{Err: errors.New("redis error")},
			wantErr:       true,
		},
		{
			name:    "market buy order is submitted",
			orderId: 3,
			mockOrderBook: mocks.MockOrderBook{
				Orders: []*models.Order{&order2LimitSell},
				Order:  order3MarketBuy,
			},
			mockExecQueue: mocks.MockExecQueue{Err: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine.OrderBook = &tt.mockOrderBook
			engine.ExecutionQueue = &tt.mockExecQueue

			gotErr := engine.SubmitOrder(ctx, tt.orderId)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("SubmitOrder() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("SubmitOrder() succeeded unexpectedly")
			}
		})
	}
}

func TestHandleIocOrder(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		incoming             *models.Order
		claimedOrders        *[]*core.ClaimedCandidate
		executionBuffer      *[]*models.ExecutionRecord
		wantStatus           string
		wantRemainingQty     int
		wantExecBufferLength int
	}{
		{
			name:                 "ioc order remaining qty is 0",
			incoming:             &models.Order{Status: "filled", Timing: "ioc", RemainingQuantity: 0},
			claimedOrders:        &[]*core.ClaimedCandidate{},
			executionBuffer:      &[]*models.ExecutionRecord{{}},
			wantStatus:           "filled",
			wantRemainingQty:     0,
			wantExecBufferLength: 1,
		},
		{
			name:                 "ioc order remaining qty is 5",
			incoming:             &models.Order{Status: "filled", Timing: "ioc", RemainingQuantity: 5, Quantity: 10},
			claimedOrders:        &[]*core.ClaimedCandidate{},
			executionBuffer:      &[]*models.ExecutionRecord{{}},
			wantStatus:           "canceled",
			wantRemainingQty:     10,
			wantExecBufferLength: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core.HandleIocOrder(tt.incoming, tt.claimedOrders, tt.executionBuffer)
			if tt.incoming.Status != tt.wantStatus {
				t.Error("wrong status")
			}
			if tt.incoming.RemainingQuantity != tt.wantRemainingQty {
				t.Error("wrong remaining qty")
			}
			if len(*tt.executionBuffer) != tt.wantExecBufferLength {
				t.Error("wrong exec buffer length")
			}
		})
	}
}

func TestRevertClaimedOrders(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		claimedOrders    *[]*core.ClaimedCandidate
		wantRemainingQty int
		wantStatus       string
	}{
		{
			name: "filled order is reversed",
			claimedOrders: &[]*core.ClaimedCandidate{
				{Order: &order4MarketBuy, ClaimedQty: 5},
			},
			wantRemainingQty: 5,
			wantStatus:       "partially_filled",
		},
		{
			name: "partially filled order is reversed",
			claimedOrders: &[]*core.ClaimedCandidate{
				{Order: &order1LimitBuy, ClaimedQty: 3},
			},
			wantRemainingQty: order4MarketBuy.Quantity,
			wantStatus:       "open",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core.RevertClaimedOrders(tt.claimedOrders)
			first := (*tt.claimedOrders)[0]
			if first.Order.RemainingQuantity != tt.wantRemainingQty {
				t.Errorf("invalid remaining qty got : %v, want: %v", first.Order.RemainingQuantity, tt.wantRemainingQty)
			}
			if first.Order.Status != tt.wantStatus {
				t.Errorf("invalid status got : %v, want: %v", first.Order.Status, tt.wantStatus)
			}
		})
	}
}
