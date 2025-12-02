package rules

import (
	"brokerx/order-service/common"
	"brokerx/order-service/models"
	"context"
	"fmt"
)

const INACTIVE_INSTRUMENT_MSG string = "The instrument is not active"

type CR002InstrumentValidity struct {
	instrument *models.Instrument
}

func (c *CR002InstrumentValidity) Setup(inputs ComplianceRuleInputs) error {
	if inputs.Instrument == nil {
		return fmt.Errorf("%w: instrument cannot be absent", common.ErrDependencyFailure)
	}
	c.instrument = inputs.Instrument
	return nil
}

func NewCR002InstrumentValidity() *CR002InstrumentValidity {
	return &CR002InstrumentValidity{}
}

func (c *CR002InstrumentValidity) Verify(ctx context.Context, order *models.Order) error {
	if c.instrument.Status != "Active" {
		return fmt.Errorf("%w: error when verifying instrument {%v}: %v", common.ErrBusinessRuleViolation, order.Symbol, INACTIVE_INSTRUMENT_MSG)
	}
	return nil
}

var _ ComplianceRule = (*CR002InstrumentValidity)(nil) // Ensure interface is implemented at compile time
