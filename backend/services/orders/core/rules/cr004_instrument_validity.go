package rules

import (
	"brokerx/order-service/models"
	"fmt"
)

const INACTIVE_INSTRUMENT_MSG string = "The instrument is not active"

type CR004InstrumentValidity struct {
	instrument *models.Instrument
}

func (c *CR004InstrumentValidity) Setup(inputs ComplianceRuleInputs) error {
	c.instrument = inputs.Instrument
	return nil
}

func NewCR004InstrumentValidity() *CR004InstrumentValidity {
	return &CR004InstrumentValidity{}
}

func (c *CR004InstrumentValidity) Verify(order *models.Order) error {
	if c.instrument.Status != "Active" {
		return fmt.Errorf("error when verifying instrument {%v}: %v", order.Symbol, INACTIVE_INSTRUMENT_MSG)
	}
	return nil
}

var _ ComplianceRule = (*CR004InstrumentValidity)(nil) // Ensure interface is implemented at compile time
