package rules

import (
	"brokerx/order-service/common"
	"brokerx/order-service/models"
	"context"
	"fmt"
)

const INVALID_PRICE_BAND_MSG string = "The specified price is outside the instrument's price band"

type CR006InstrumentPriceBand struct {
	instrument   *models.Instrument
	currentPrice float64
}

func (c *CR006InstrumentPriceBand) Setup(inputs ComplianceRuleInputs) error {
	if inputs.CurrentPrice == nil || inputs.Instrument == nil {
		return fmt.Errorf("%w: neither instrument nor currenPrice can be absent", common.ErrDependencyFailure)
	}
	c.instrument = inputs.Instrument
	c.currentPrice = *inputs.CurrentPrice
	return nil
}

func NewCR006InstrumentPriceBand() *CR006InstrumentPriceBand {
	return &CR006InstrumentPriceBand{}
}

func (c *CR006InstrumentPriceBand) Verify(ctx context.Context, order *models.Order) error {
	if order.Type == "market" {
		return nil
	}
	
	maxPrice := c.currentPrice + c.currentPrice*(1.0/float64(c.instrument.PriceBandPercent))
	minPrice := c.currentPrice - c.currentPrice*(1.0/float64(c.instrument.PriceBandPercent))
	if order.UnitPrice < minPrice || order.UnitPrice > maxPrice {
		return fmt.Errorf("%w: error when validating price band {%v}: %v", common.ErrBusinessRuleViolation, order.Symbol, INVALID_PRICE_BAND_MSG)
	}

	return nil
}

var _ ComplianceRule = (*CR006InstrumentPriceBand)(nil) // Ensure interface is implemented at compile time
