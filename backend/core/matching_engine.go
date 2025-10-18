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
	return engine.TransactionManager.Do(context.Background(), func(orders ports.OrderRepository, executions ports.ExecutionRepository) error {
		var (
			matchedOrders []*models.Order
			err           error
		)

		switch order.Type {
		case "market":
			matchedOrders, err = orders.FindMatchesMarket(order)

		case "limit":
			matchedOrders, err = orders.FindMatchesLimit(order, order.UnitPrice)
		}
		if err != nil {
			log.Errorf("Error when fetching orders: %v", err)
			return err
		}

		chosenOrders := make([]*models.Order, 0, len(matchedOrders))
		executionRecords := make([]*models.ExecutionRecord, 0, len(matchedOrders))
		for _, m := range matchedOrders {
			if order.RemainingQuantity == 0 {
				break
			}

			qty := min(order.RemainingQuantity, m.RemainingQuantity)
			order.RemainingQuantity -= qty
			m.RemainingQuantity -= qty

			m = updateStatus(m)
			m = updateMatchedOrderUnitPrice(m, qty, order.UnitPrice)
			chosenOrders = append(chosenOrders, m)

			exec := &models.ExecutionRecord{
				BuyOrderID:  pickID(order, m, "buy"),
				SellOrderID: pickID(order, m, "sell"),
				Symbol:      order.Symbol,
				Price:       m.UnitPrice,
				Quantity:    int(qty),
			}
			executionRecords = append(executionRecords, exec)
		}

		if order.Timing == "ioc" && order.RemainingQuantity > 0 {
			order.Status = "canceled"
			if err := orders.Update(order); err != nil {
				log.Errorf("Error when updating order #%d: %v", order.ID, err)
				return err
			}
			return nil
		}
		order = updateStatus(order)
		order = updateUnitPrice(order, executionRecords)
		if err := orders.Update(order); err != nil {
			log.Errorf("Error when updating order #%d: %v", order.ID, err)
			return err
		}

		for i, o := range chosenOrders {
			if err := orders.Update(o); err != nil {
				log.Errorf("Error when updating matched order #%d: %v", o.ID, err)
				return err
			}

			if err := executions.Create(executionRecords[i]); err != nil {
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

func updateUnitPrice(order *models.Order, execRecords []*models.ExecutionRecord) *models.Order {
	if order.Type == "limit" {
		return order
	}

	var totalQty int
	var totalValue float64
	for _, e := range execRecords {
		totalQty += e.Quantity
		totalValue += float64(e.Quantity) * e.Price
	}
	weightedAvg := totalValue / float64(totalQty)

	order.UnitPrice = weightedAvg
	return order
}

func updateMatchedOrderUnitPrice(order *models.Order, qty int, original_order_unit_price float64) *models.Order {
	if order.Type == "limit" {
		return order
	}

	totalQty := order.Quantity - order.RemainingQuantity
	totalValue := ((float64(totalQty) - float64(qty)) * order.UnitPrice) + (float64(qty) * original_order_unit_price)

	order.UnitPrice = totalValue / float64(totalQty)
	return order
}

func pickID(a, b *models.Order, side string) int {
	if a.Action == side {
		return a.ID
	}
	return b.ID
}

var _ ports.MatchingEngine = (*MatchingEngine)(nil) // Ensure interface is implemented at compile time
