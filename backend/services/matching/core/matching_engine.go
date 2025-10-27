package core

import (
	"brokerx/matching-service/models"
	"brokerx/matching-service/ports"

	log "github.com/sirupsen/logrus"
)

type MatchingEngine interface {
	SubmitOrder(orderId int) error
}

type MatchingEngineImpl struct {
	OrderBook      ports.OrderBook
	ExecutionQueue ports.ExecutionQueue
}

type claimedCandidate struct {
	Order      *models.Order
	ClaimedQty int
}

func (engine *MatchingEngineImpl) SubmitOrder(orderId int) error {
	// 1. Fetch order from order book
	order, err := engine.OrderBook.GetById(orderId)
	if err != nil {
		return err
	}
	if order == (models.Order{}) {
		log.Infof("Order #%d doesn't exist or is already being processed as a candidate.", orderId)
		return nil
	}

	var (
		allMatchedOrders []*models.Order
		claimedOrders    []*claimedCandidate
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
			matchedOrders, err = engine.OrderBook.FindMatchesMarket(order.Symbol, order.Type, order.Action, batchSize)
		case "limit":
			matchedOrders, err = engine.OrderBook.FindMatchesLimit(order.Symbol, order.Type, order.Action, order.UnitPrice, batchSize)
		}
		if err != nil {
			log.Errorf("Error when fetching candidate matches: %v", err)
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

			updateStatus(candidate)

			claimedOrders = append(claimedOrders, &claimedCandidate{
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

	updateStatus(&order)
	handleIocOrder(&order, &claimedOrders, &executionBuffer)

	// Returns incomplete orders to order book
	if err := engine.handleRemainingOrders(order, &allMatchedOrders); err != nil {
		return err
	}

	// Send complete incoming order and candidates to queue for persistence
	if err := engine.saveOrdersAndExecutions(&order, claimedOrders, executionBuffer); err != nil {
		return err
	}

	return nil
}

func handleIocOrder(incoming *models.Order, claimedOrders *[]*claimedCandidate, executionBuffer *[]*models.ExecutionRecord) {
	if incoming.Timing != "ioc" {
		return
	}
	log.Infof("handle finalization of ioc order : %v", incoming)

	if incoming.RemainingQuantity > 0 {
		incoming.Status = "canceled"
		incoming.RemainingQuantity = incoming.Quantity
		revertClaimedOrders(claimedOrders)
		*executionBuffer = nil
	}
}

func revertClaimedOrders(claimedOrders *[]*claimedCandidate) {
	for _, claimed := range *claimedOrders {
		claimed.Order.RemainingQuantity += claimed.ClaimedQty
		claimed.Order = updateStatus(claimed.Order)
	}
}

func (engine *MatchingEngineImpl) handleRemainingOrders(incoming models.Order, allMatchedOrders *[]*models.Order) error {
	ordersToReturn := []*models.Order{}

	if incoming.Status == "partially_filled" || incoming.Status == "open" {
		ordersToReturn = append(ordersToReturn, &incoming)
	}

	for _, matched := range *allMatchedOrders {
		if matched.Status == "partially_filled" || matched.Status == "open" {
			ordersToReturn = append(ordersToReturn, matched)
		}
	}

	log.Infof("length of orders to return : %d", len(ordersToReturn))

	return engine.OrderBook.Return(ordersToReturn)
}

func (engine *MatchingEngineImpl) saveOrdersAndExecutions(incoming *models.Order, claimedOrders []*claimedCandidate, executionBuffer []*models.ExecutionRecord) error {
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
		if err := engine.OrderBook.EnqueueOrders(ordersToPersist); err != nil {
			log.Errorf("error when enqueuing orders for saving to db")
		}
	}

	if len(executionBuffer) > 0 {
		if err := engine.ExecutionQueue.EnqueueExecutionRecords(executionBuffer); err != nil {
			log.Errorf("error when enqueuing execution records for saving to db")
		}
	}

	return nil
}

func updateStatus(order *models.Order) *models.Order {
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
