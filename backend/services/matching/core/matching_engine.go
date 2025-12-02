package core

import (
	"brokerx/matching-service/models"
	"brokerx/matching-service/ports"
	"brokerx/matching-service/util"
	"context"
	"log/slog"
)

type QueuedOrder struct {
    Ctx   context.Context
    Order *models.Order
}

var OrderQueue chan QueuedOrder

type MatchingEngine interface {
	SubmitOrder(ctx context.Context, order *models.Order) error
}

type MatchingEngineImpl struct {
	OrderBook      ports.OrderBook
	Producer 	   ports.EventProducer
}

func (engine *MatchingEngineImpl) SubmitOrder(ctx context.Context, order *models.Order) error {
	log := util.FromContext(ctx)

	var (
		usedOrders 			[]*models.Order
		unclaimedCandidates []*models.Order
		claimedOrders    []*models.ClaimedCandidate
		batchSize        = 5
		executionBuffer  []*models.ExecutionRecord
	)

	for order.RemainingQuantity > 0 {
		var (
			matchedOrders []*models.Order
			err           error
		)

		// 1. Fetch candidate matches
		matchedOrders, err = engine.OrderBook.FindMatches(ctx, order, batchSize)
		if err != nil {
			slog.Error("Error when fetching candidate matches", "orderId", order.ID, "error", err)
			return err
		}
		if len(matchedOrders) == 0 {
			break
		}

		// 2. Try matching each candidate
		for _, candidate := range matchedOrders {
			candidateCopy := candidate
			usedOrders = append(usedOrders, candidateCopy)

			if order.RemainingQuantity == 0 {
				unclaimedCandidates = append(unclaimedCandidates, candidate)
				continue
			}

			qty := min(order.RemainingQuantity, candidate.RemainingQuantity)
			order.RemainingQuantity -= qty
			candidate.RemainingQuantity -= qty

			UpdateStatus(candidate)

			claimedOrders = append(claimedOrders, &models.ClaimedCandidate{
				Order:      *candidate,
				ClaimedQty: qty,
			})

			executionBuffer = append(executionBuffer, &models.ExecutionRecord{
				BuyOrderID:  pickID(order, candidate, "buy"),
				SellOrderID: pickID(order, candidate, "sell"),
				Symbol:      order.Symbol,
				Price:       pickUnitPrice(order, candidate),
				Quantity:    qty,
			})
		}
	}

	UpdateStatus(order)
	HandleIocOrder(order, &claimedOrders, &unclaimedCandidates, &executionBuffer)

	engine.OrderBook.Return(ctx, unclaimedCandidates)

	if order.Timing == "ioc" && order.Status == "canceled" {
		return engine.Producer.SendEvent(ctx, "OrderEvents", "OrderMatchingFailed", *order, nil)
	}

	if err := engine.Producer.SendMatchingEvent(ctx, "PortfolioEvents", "OrderMatched", *order, claimedOrders, executionBuffer, nil); err != nil {
		log.Error("error sending OrderMatched event. returning all used orders...", "error", err)
		engine.OrderBook.Return(ctx, usedOrders)
		return err
	}

	return nil
}

func HandleIocOrder(incoming *models.Order, claimedOrders *[]*models.ClaimedCandidate, unclaimedCandidates *[]*models.Order, executionBuffer *[]*models.ExecutionRecord) {
	if incoming.Timing != "ioc" {
		return
	}

	if incoming.RemainingQuantity > 0 {
		incoming.Status = "canceled"
		incoming.RemainingQuantity = incoming.Quantity
		RevertClaimedOrders(claimedOrders, unclaimedCandidates)
		*executionBuffer = nil
	}
}

func RevertClaimedOrders(claimedOrders *[]*models.ClaimedCandidate, unclaimedCandidates *[]*models.Order) {
	for _, claimed := range *claimedOrders {
		claimed.Order.RemainingQuantity += claimed.ClaimedQty
		claimed.Order = *UpdateStatus(&claimed.Order)
		*unclaimedCandidates = append(*unclaimedCandidates, &claimed.Order)
	}
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
