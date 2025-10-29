package rules

import (
	"brokerx/order-service/common"
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"context"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type CR002BuyingPower struct {
	PortfolioService ports.PortfolioService
	currentPrice     float64
}

func (c *CR002BuyingPower) Setup(inputs ComplianceRuleInputs) error {
	if inputs.CurrentPrice == nil {
		return errors.New("currentPrice cannot be absent")
	}
	c.currentPrice = *inputs.CurrentPrice
	return nil
}

func NewCR002BuyingPower(portfolioService ports.PortfolioService) *CR002BuyingPower {
	return &CR002BuyingPower{PortfolioService: portfolioService}
}

func (c *CR002BuyingPower) Verify(ctx context.Context, order *models.Order) error {
	if order.Action == "sell" {
		return nil
	}

	wallet, err := c.PortfolioService.GetWallet(ctx, order.UserID)
	if err != nil {
		return fmt.Errorf("%w: %v", common.ErrDependencyFailure, err)
	}

	log.Infof("available funds %v", wallet.AvailableFunds)
	if wallet.AvailableFunds < (c.currentPrice * float64(order.Quantity)) {
		return fmt.Errorf("%w: not enough available funds", common.ErrBusinessRuleViolation)
	}

	return nil
}

var _ ComplianceRule = (*CR002BuyingPower)(nil) // Ensure interface is implemented at compile time
