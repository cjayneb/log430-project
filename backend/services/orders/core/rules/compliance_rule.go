package rules

import (
	"brokerx/order-service/models"
	"context"
)

type ComplianceRuleInputs struct {
	Instrument   *models.Instrument
	CurrentPrice *float64
}

type ComplianceRule interface {
	Verify(ctx context.Context, order *models.Order) error
	Setup(inputs ComplianceRuleInputs) error
}
