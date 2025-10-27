package rules

import (
	"brokerx/order-service/common"
	"brokerx/order-service/models"
	"context"
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

func (c *CR001OrderQuantity) Verify(ctx context.Context, order *models.Order) error {
	if order.Quantity < MIN_ORDER_QUANTITY {
		return fmt.Errorf("%w: order quantity must be at least %v", common.ErrBusinessRuleViolation, MIN_ORDER_QUANTITY)
	}

	if order.Quantity > MAX_ORDER_QUANTITY {
		return fmt.Errorf("%w: order quantity surpasses maximum quantity allowed by BrokerX (%v)", common.ErrBusinessRuleViolation, MAX_ORDER_QUANTITY)
	}

	return nil
}

var _ ComplianceRule = (*CR001OrderQuantity)(nil) // Ensure interface is implemented at compile time
