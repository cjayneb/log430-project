package rules

import (
	"brokerx/order-service/models"
	"fmt"
)

const INVALID_PRICE_BAND_MSG string = "The specified price is outside the instrument's price band"

type CR006InstrumentPriceBand struct {
	instrument   *models.Instrument
	currentPrice float64
}

func (c *CR006InstrumentPriceBand) Setup(inputs ComplianceRuleInputs) error {
	c.instrument = inputs.Instrument
	c.currentPrice = inputs.CurrentPrice
	return nil
}

func NewCR006InstrumentPriceBand() *CR006InstrumentPriceBand {
	return &CR006InstrumentPriceBand{}
}

func (c *CR006InstrumentPriceBand) Verify(order *models.Order) error {
	maxPrice := c.currentPrice + c.currentPrice*(1.0/float64(c.instrument.PriceBandPercent))
	minPrice := c.currentPrice - c.currentPrice*(1.0/float64(c.instrument.PriceBandPercent))
	if order.UnitPrice < minPrice || order.UnitPrice > maxPrice {
		return fmt.Errorf("error when validating price band {%v}: %v", order.Symbol, INVALID_PRICE_BAND_MSG)
	}

	return nil
}

var _ ComplianceRule = (*CR006InstrumentPriceBand)(nil) // Ensure interface is implemented at compile time
