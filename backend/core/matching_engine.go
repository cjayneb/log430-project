package core

import (
	"brokerx/models"
	"brokerx/ports"
	"context"

	log "github.com/sirupsen/logrus"
)

type MatchingEngine struct {
	TransactionManager ports.TransactionManager
}

func (engine *MatchingEngine) SubmitOrder(order *models.Order) error {
	var claimedOrders []*models.Order

	for order.RemainingQuantity > 0 {
		var (
			matchedOrders []*models.Order
			err           error
		)

		switch order.Type {
		case "market":
			matchedOrders, err = engine.fetchCandidatesMarket(order)

		case "limit":
			matchedOrders, err = engine.fetchCandidatesLimit(order)
		}
		if err != nil {
			log.Errorf("Error when fetching candidate matches: %v", err)
			return err
		}
		if len(matchedOrders) == 0 {
			break
		}

		for _, candidate := range matchedOrders {
			if candidate.UserID == order.UserID {
				continue
			}
			if order.RemainingQuantity == 0 {
				break
			}

			qty := min(order.RemainingQuantity, candidate.RemainingQuantity)
			success, err := engine.tryMatch(order, candidate, qty)
			if err != nil {
				log.Errorf("Matching error for order #%d with candidate #%d: %v", order.ID, candidate.ID, err)
				continue
			}
			if success {
				order.RemainingQuantity -= qty
				claimedOrders = append(claimedOrders, candidate)
			}
		}
	}

	if order.Timing == "ioc" && order.RemainingQuantity > 0 {
		for _, c := range claimedOrders {
			err := engine.TransactionManager.Do(context.Background(), func(orders ports.OrderRepository, _ ports.ExecutionRepository) error {
				return orders.RevertClaim(c.ID, c.UnitPrice, c.Quantity)
			})
			if err != nil {
				log.Errorf("error when reverting claim on order #%d when cancelling order #%d", c.ID, order.ID)
			}
		}

		order.Status = "canceled"
		return engine.TransactionManager.Do(context.Background(), func(orders ports.OrderRepository, exec ports.ExecutionRepository) error {
			return orders.Update(order)
		})
	}

	return nil
}

func (engine *MatchingEngine) fetchCandidatesLimit(order *models.Order) ([]*models.Order, error) {
	return engine.TransactionManager.DoReadOnly(context.Background(), func(orders ports.OrderRepository) ([]*models.Order, error) {
		return orders.FindMatchesLimit(order, order.UnitPrice, 10)
	})
}

func (engine *MatchingEngine) fetchCandidatesMarket(order *models.Order) ([]*models.Order, error) {
	return engine.TransactionManager.DoReadOnly(context.Background(), func(orders ports.OrderRepository) ([]*models.Order, error) {
		return orders.FindMatchesMarket(order, 10)
	})
}

func (engine *MatchingEngine) tryMatch(incoming *models.Order, candidate *models.Order, qty int) (bool, error) {
	returnValue := false

	err := engine.TransactionManager.Do(context.Background(), func(orders ports.OrderRepository, executions ports.ExecutionRepository) error {
		// 1. Try to atomically claim qty from the candidate order
		affected, err := orders.ClaimOrder(candidate.ID, determineUnitPrice(candidate, qty, incoming.UnitPrice), qty)
		if err != nil {
			return err
		}
		if affected == 0 {
			// Someone else matched this candidate already
			return nil
		}

		// 2. Record execution
		exec := &models.ExecutionRecord{
			BuyOrderID:  pickID(incoming, candidate, "buy"),
			SellOrderID: pickID(incoming, candidate, "sell"),
			Symbol:      incoming.Symbol,
			Price:       pickUnitPrice(incoming, candidate),
			Quantity:    qty,
		}
		if err := executions.Create(exec); err != nil {
			return err
		}

		// 3. Update the incoming order’s status in DB (partial fill or filled)
		incoming.RemainingQuantity -= qty
		incoming = updateStatus(incoming)
		incoming.UnitPrice = determineUnitPrice(incoming, qty, candidate.UnitPrice)
		if err := orders.Update(incoming); err != nil {
			return err
		}

		returnValue = true
		return nil
	})
	return returnValue, err
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
