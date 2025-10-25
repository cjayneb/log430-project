package core

import (
	"brokerx/models"
	"brokerx/ports"
	"context"

	log "github.com/sirupsen/logrus"
)

type MatchingEngine struct {
	TransactionManager ports.TransactionManager
	OrderBook          ports.OrderBook
}

type claimedCandidate struct {
	Order      *models.Order
	ClaimedQty int
}

func (engine *MatchingEngine) SubmitOrder(orderId int) error {
	// 1. Fetch order from order book
	order, err := engine.OrderBook.GetById(orderId)
	if err != nil {
		return err
	}
	if order == (models.Order{}) {
		log.Infof("Order #%d is already being processed as a candidate.", orderId)
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
			log.Infof("candidate : %v", candidate)
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
	handleIocOrder(&order, &claimedOrders)

	// Returns incomplete orders to order book
	if err := engine.handleRemainingOrders(order, &allMatchedOrders); err != nil {
		return err
	}

	// Persist incoming order and candidates that are complete
	return engine.persistOrdersAndExecutions(&order, claimedOrders, executionBuffer)
}

func handleIocOrder(incoming *models.Order, claimedOrders *[]*claimedCandidate) {
	if incoming.Timing != "ioc" {
		return
	}

	if incoming.RemainingQuantity > 0 {
		incoming.Status = "canceled"
		incoming.RemainingQuantity = incoming.Quantity
		revertClaimedOrders(claimedOrders)
	}
}

func revertClaimedOrders(claimedOrders *[]*claimedCandidate) {
	for _, claimed := range *claimedOrders {
		claimed.Order.RemainingQuantity -= claimed.ClaimedQty
		claimed.Order = updateStatus(claimed.Order)
	}
}

func (engine *MatchingEngine) handleRemainingOrders(incoming models.Order, allMatchedOrders *[]*models.Order) error {
	ordersToReturn := []*models.Order{}

	if incoming.Status == "partially_filled" || incoming.Status == "open" {
		ordersToReturn = append(ordersToReturn, &incoming)
	}

	for _, matched := range *allMatchedOrders {
		if matched.Status == "partially_filled" || matched.Status == "open" {
			ordersToReturn = append(ordersToReturn, matched)
		}
	}

	return engine.OrderBook.Return(ordersToReturn)
}

func (engine *MatchingEngine) persistOrdersAndExecutions(incoming *models.Order, claimedOrders []*claimedCandidate, executionBuffer []*models.ExecutionRecord) error {
	return engine.TransactionManager.Do(context.Background(), func(orders ports.OrderRepository, executions ports.ExecutionRepository) error {
		ordersToPersist := []*models.Order{}
		if incoming.Status == "filled" || incoming.Status == "canceled" {
			ordersToPersist = append(ordersToPersist, incoming)
		}

		for _, claimed := range claimedOrders {
			if claimed.Order.Status == "filled" {
				ordersToPersist = append(ordersToPersist, claimed.Order)
			}
		}

		// TODO: send orders to queue and flush every x seconds to reduce IO
		if err := orders.UpdateBatch(ordersToPersist); err != nil {
			log.Errorf("error saving orders: %v", err)
			return err
		}

		// TODO: send executions to queue and flush every x seconds to reduce IO
		if len(executionBuffer) > 0 {
			if err := executions.CreateBatch(executionBuffer); err != nil {
				log.Errorf("error flushing execution buffer: %v", err)
				return err
			}
		}

		return nil
	})
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

	return order
}

func determineUnitPrice(order *models.Order, qty int, otherOrderUnitPrice float64) float64 {
	if order.Type == "limit" {
		return order.UnitPrice
	}

	totalQty := order.Quantity - order.RemainingQuantity
	totalValue := (float64(totalQty) * order.UnitPrice) + (float64(qty) * otherOrderUnitPrice)

	return totalValue / float64(totalQty+qty)
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

var _ ports.MatchingEngine = (*MatchingEngine)(nil) // Ensure interface is implemented at compile time
