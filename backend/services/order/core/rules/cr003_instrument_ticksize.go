package rules

import (
	"brokerx/order-service/common"
	"brokerx/order-service/models"
	"context"
	"fmt"
	"math"
)

const INVALID_TICK_SIZE_MSG string = "The specified price does not match the instrument's tick size"

type CR003InstrumentTickSize struct {
	instrument *models.Instrument
}

func (c *CR003InstrumentTickSize) Setup(inputs ComplianceRuleInputs) error {
	if inputs.Instrument == nil {
		return fmt.Errorf("%w: instrument cannot be absent", common.ErrDependencyFailure)
	}
	c.instrument = inputs.Instrument
	return nil
}

func NewCR003InstrumentTickSize() *CR003InstrumentTickSize {
	return &CR003InstrumentTickSize{}
}

func (c *CR003InstrumentTickSize) Verify(ctx context.Context, order *models.Order) error {
	if order.Type == "market" {
		return nil
	}

	if !isValidTick(order.UnitPrice, c.instrument.TickSize) {
		return fmt.Errorf("%w: error when validating tick size {%v}: %v", common.ErrBusinessRuleViolation, order.Symbol, INVALID_TICK_SIZE_MSG)
	}

	return nil
}

func isValidTick(price, tickSize float64) bool {
	remainder := math.Mod(price, tickSize)
	return remainder < 1e-9 || math.Abs(remainder-tickSize) < 1e-9
}

var _ ComplianceRule = (*CR003InstrumentTickSize)(nil) // Ensure interface is implemented at compile time
