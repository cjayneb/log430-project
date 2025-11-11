package core

import (
	"brokerx/matching-service/models"
	"brokerx/matching-service/ports"
	"brokerx/matching-service/util"
	"context"
	"fmt"
	"log/slog"
)

type QueuedOrder struct {
    Ctx   context.Context
    Order *models.Order
}

var OrderQueue chan QueuedOrder

type MatchingEngine interface {
	QueueOrder(ctx context.Context, order *models.Order) error
}

type MatchingEngineImpl struct {
	OrderBook      ports.OrderBook
	ExecutionQueue ports.ExecutionQueue
}

type ClaimedCandidate struct {
	Order      *models.Order
	ClaimedQty int
}

func (engine *MatchingEngineImpl) QueueOrder(ctx context.Context, order *models.Order) error {
	log := util.FromContext(ctx)

	orderQueueLen.Set(float64(len(OrderQueue)))
	OrderQueue <- QueuedOrder{Ctx: context.WithoutCancel(ctx), Order: order}
	log.Info("Order queued for matching", "orderId", order.ID)
	return nil
}

func (service *MatchingEngineImpl) StartMatchingWorkers(numberOfGoRoutines int) {
	OrderQueue = make(chan QueuedOrder, 1000)
	for i := 0; i < numberOfGoRoutines; i++ {
		go func() {
			for qo := range OrderQueue {
				log := util.FromContext(qo.Ctx)
				if err := service.OrderBook.Insert(qo.Ctx, qo.Order); err != nil {
					// TODO: retry? or find a way to let user know
					log.Error("Order book insertion failed for order", "orderId", qo.Order.ID, "error", err)
				}
				if err := service.SubmitOrder(qo.Ctx, qo.Order.ID); err != nil {
					log.Error("matching failed for order", "orderId", qo.Order.ID, "error", err)
				}
			}
		}()
	}
	slog.Info(fmt.Sprintf("Started %d matching workers", numberOfGoRoutines))
}

func (engine *MatchingEngineImpl) SubmitOrder(ctx context.Context, orderId int) error {
	// 1. Fetch order from order book
	order, err := engine.OrderBook.GetById(ctx, orderId)
	if err != nil {
		return err
	}
	if order == (models.Order{}) {
		slog.Info("Order doesn't exist or is already being processed as a candidate", "orderId", orderId)
		return nil
	}

	var (
		allMatchedOrders []*models.Order
		claimedOrders    []*ClaimedCandidate
		batchSize        = 5
		executionBuffer  []*models.ExecutionRecord
	)

	for order.RemainingQuantity > 0 {
		var (
			matchedOrders []*models.Order
			err           error
		)

		// 2. Fetch candidate matches
		switch order.Type {
		case "market":
			matchedOrders, err = engine.OrderBook.FindMatchesMarket(ctx, order.Symbol, order.Action, batchSize)
		case "limit":
			matchedOrders, err = engine.OrderBook.FindMatchesLimit(ctx, order.Symbol, order.Action, order.UnitPrice, batchSize)
		}
		if err != nil {
			slog.Error("Error when fetching candidate matches", "orderId", orderId, "error", err)
			return err
		}
		if len(matchedOrders) == 0 {
			break
		}

		// 3. Try matching each candidate
		for _, candidate := range matchedOrders {
			// TODO: remove this filtering and filter at order book level
			if candidate.UserID == order.UserID {
				continue
			}
			if order.RemainingQuantity == 0 {
				break
			}

			qty := min(order.RemainingQuantity, candidate.RemainingQuantity)
			order.RemainingQuantity -= qty
			candidate.RemainingQuantity -= qty

			UpdateStatus(candidate)

			claimedOrders = append(claimedOrders, &ClaimedCandidate{
				Order:      candidate,
				ClaimedQty: qty,
			})

			executionBuffer = append(executionBuffer, &models.ExecutionRecord{
				BuyOrderID:  pickID(&order, candidate, "buy"),
				SellOrderID: pickID(&order, candidate, "sell"),
				Symbol:      order.Symbol,
				Price:       pickUnitPrice(&order, candidate),
				Quantity:    qty,
			})
		}
		allMatchedOrders = append(allMatchedOrders, matchedOrders...)
	}

	UpdateStatus(&order)
	HandleIocOrder(&order, &claimedOrders, &executionBuffer)

	// Returns incomplete orders to order book
	if err := engine.handleRemainingOrders(ctx, order, &allMatchedOrders); err != nil {
		return err
	}

	// Send complete incoming order and candidates to queue for persistence
	if err := engine.saveOrdersAndExecutions(ctx, &order, claimedOrders, executionBuffer); err != nil {
		return err
	}

	return nil
}

func HandleIocOrder(incoming *models.Order, claimedOrders *[]*ClaimedCandidate, executionBuffer *[]*models.ExecutionRecord) {
	if incoming.Timing != "ioc" {
		return
	}

	if incoming.RemainingQuantity > 0 {
		incoming.Status = "canceled"
		incoming.RemainingQuantity = incoming.Quantity
		RevertClaimedOrders(claimedOrders)
		*executionBuffer = nil
	}
}

func RevertClaimedOrders(claimedOrders *[]*ClaimedCandidate) {
	for _, claimed := range *claimedOrders {
		claimed.Order.RemainingQuantity += claimed.ClaimedQty
		claimed.Order = UpdateStatus(claimed.Order)
	}
}

func (engine *MatchingEngineImpl) handleRemainingOrders(ctx context.Context, incoming models.Order, allMatchedOrders *[]*models.Order) error {
	ordersToReturn := []*models.Order{}

	if incoming.Status == "partially_filled" || incoming.Status == "open" {
		ordersToReturn = append(ordersToReturn, &incoming)
	}

	for _, matched := range *allMatchedOrders {
		if matched.Status == "partially_filled" || matched.Status == "open" {
			ordersToReturn = append(ordersToReturn, matched)
		}
	}

	return engine.OrderBook.Return(ctx, ordersToReturn)
}

func (engine *MatchingEngineImpl) saveOrdersAndExecutions(ctx context.Context, incoming *models.Order, claimedOrders []*ClaimedCandidate, executionBuffer []*models.ExecutionRecord) error {
	ordersToPersist := []*models.Order{}
	if incoming.Status == "filled" || incoming.Status == "canceled" {
		ordersToPersist = append(ordersToPersist, incoming)
	}

	for _, claimed := range claimedOrders {
		if claimed.Order.Status == "filled" {
			ordersToPersist = append(ordersToPersist, claimed.Order)
		}
	}

	if len(ordersToPersist) > 0 {
		if err := engine.OrderBook.EnqueueOrders(ctx, ordersToPersist); err != nil {
			slog.Error("error when enqueuing orders for saving to db", "error", err)
		}
	}

	if len(executionBuffer) > 0 {
		if err := engine.ExecutionQueue.EnqueueExecutionRecords(ctx, executionBuffer); err != nil {
			slog.Error("error when enqueuing execution records for saving to db", "error", err)
		}
	}

	return nil
}

func UpdateStatus(order *models.Order) *models.Order {
	if order.RemainingQuantity != 0 && order.RemainingQuantity < order.Quantity {
		order.Status = "partially_filled"
		return order
	}

	if order.RemainingQuantity == 0 {
		order.Status = "filled"
		return order
	}

	if order.RemainingQuantity == order.Quantity {
		order.Status = "open"
		return order
	}

	return order
}

func pickID(a, b *models.Order, side string) int {
	if a.Action == side {
		return a.ID
	}
	return b.ID
}

func pickUnitPrice(incoming, candidate *models.Order) float64 {
	if incoming.Type == "market" || candidate.Type == "limit" {
		return candidate.UnitPrice
	}
	return incoming.UnitPrice
}

var _ MatchingEngine = (*MatchingEngineImpl)(nil) // Ensure interface is implemented at compile time
