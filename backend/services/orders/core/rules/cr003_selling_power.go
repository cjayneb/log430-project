package rules

import (
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"errors"
)

type CR003SellingPower struct {
	PortfolioService ports.PortfolioService
}

func (c *CR003SellingPower) Setup(inputs ComplianceRuleInputs) error {
	return nil
}

func NewCR003SellingPower(portfolioService ports.PortfolioService) *CR003SellingPower {
	return &CR003SellingPower{PortfolioService: portfolioService}
}

func (c *CR003SellingPower) Verify(order *models.Order) error {
	if order.Action == "buy" {
		return nil
	}

	positions, err := c.PortfolioService.FetchPositions(order.UserID, order.Symbol)
	if err != nil {
		return err
	}

	totalOwnedStock := 0
	for _, p := range positions {
		totalOwnedStock += p.Quantity
	}
	if totalOwnedStock < order.Quantity {
		return errors.New("not enough owned stocks")
	}

	return nil
}

var _ ComplianceRule = (*CR003SellingPower)(nil) // Ensure interface is implemented at compile time
