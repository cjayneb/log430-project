package rules

import (
	"brokerx/order-service/models"
)

type ComplianceRuleInputs struct {
	Instrument   *models.Instrument
	CurrentPrice float64
}

type ComplianceRule interface {
	Verify(order *models.Order) error
	Setup(inputs ComplianceRuleInputs) error
}
