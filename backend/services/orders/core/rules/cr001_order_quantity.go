package rules

import (
	"brokerx/order-service/models"
	"fmt"
)

const MAX_ORDER_QUANTITY int = 100000
const MIN_ORDER_QUANTITY int = 1

type CR001OrderQuantity struct{}

func (c *CR001OrderQuantity) Setup(inputs ComplianceRuleInputs) error {
	return nil
}

func NewCR001OrderQuantity() *CR001OrderQuantity {
	return &CR001OrderQuantity{}
}

func (c *CR001OrderQuantity) Verify(order *models.Order) error {
	if order.Quantity < MIN_ORDER_QUANTITY {
		return fmt.Errorf("order quantity must be at least %v", MIN_ORDER_QUANTITY)
	}

	if order.Quantity > MAX_ORDER_QUANTITY {
		return fmt.Errorf("order quantity surpasses maximum quantity allowed by BrokerX (%v)", MAX_ORDER_QUANTITY)
	}

	return nil
}

var _ ComplianceRule = (*CR001OrderQuantity)(nil) // Ensure interface is implemented at compile time
