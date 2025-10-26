package rules

import (
	"brokerx/order-service/models"
	"fmt"
	"math"
)

const INVALID_TICK_SIZE_MSG string = "The specified price does not match the instrument's tick size"

type C005InstrumentTickSize struct {
	instrument *models.Instrument
}

func (c *C005InstrumentTickSize) Setup(inputs ComplianceRuleInputs) error {
	c.instrument = inputs.Instrument
	return nil
}

func NewC005InstrumentTickSize() *C005InstrumentTickSize {
	return &C005InstrumentTickSize{}
}

func (c *C005InstrumentTickSize) Verify(order *models.Order) error {
	if order.Type == "market" {
		return nil
	}

	if !isValidTick(order.UnitPrice, c.instrument.TickSize) {
		return fmt.Errorf("error when validating tick size {%v}: %v", order.Symbol, INVALID_TICK_SIZE_MSG)
	}

	return nil
}

func isValidTick(price, tickSize float64) bool {
	remainder := math.Mod(price, tickSize)
	return remainder < 1e-9 || math.Abs(remainder-tickSize) < 1e-9
}

var _ ComplianceRule = (*C005InstrumentTickSize)(nil) // Ensure interface is implemented at compile time
