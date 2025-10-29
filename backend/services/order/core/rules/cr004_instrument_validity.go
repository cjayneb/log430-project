package rules

import (
	"brokerx/order-service/common"
	"brokerx/order-service/models"
	"context"
	"fmt"
)

const INACTIVE_INSTRUMENT_MSG string = "The instrument is not active"

type CR004InstrumentValidity struct {
	instrument *models.Instrument
}

func (c *CR004InstrumentValidity) Setup(inputs ComplianceRuleInputs) error {
	if inputs.Instrument == nil {
		return fmt.Errorf("%w: instrument cannot be absent", common.ErrDependencyFailure)
	}
	c.instrument = inputs.Instrument
	return nil
}

func NewCR004InstrumentValidity() *CR004InstrumentValidity {
	return &CR004InstrumentValidity{}
}

func (c *CR004InstrumentValidity) Verify(ctx context.Context, order *models.Order) error {
	if c.instrument.Status != "Active" {
		return fmt.Errorf("%w: error when verifying instrument {%v}: %v", common.ErrBusinessRuleViolation, order.Symbol, INACTIVE_INSTRUMENT_MSG)
	}
	return nil
}

var _ ComplianceRule = (*CR004InstrumentValidity)(nil) // Ensure interface is implemented at compile time
