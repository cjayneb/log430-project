package rules

import (
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"errors"
)

type CR002BuyingPower struct {
	PortfolioService ports.PortfolioService
	currentPrice     float64
}

func (c *CR002BuyingPower) Setup(inputs ComplianceRuleInputs) error {
	c.currentPrice = inputs.CurrentPrice
	return nil
}

func NewCR002BuyingPower(portfolioService ports.PortfolioService) *CR002BuyingPower {
	return &CR002BuyingPower{PortfolioService: portfolioService}
}

func (c *CR002BuyingPower) Verify(order *models.Order) error {
	if order.Action == "sell" {
		return nil
	}

	wallet, err := c.PortfolioService.GetWallet(order.UserID)
	if err != nil {
		return err
	}

	if wallet.AvailableFunds < (c.currentPrice * float64(order.Quantity)) {
		return errors.New("not enough available funds")
	}

	return nil
}

var _ ComplianceRule = (*CR002BuyingPower)(nil) // Ensure interface is implemented at compile time
