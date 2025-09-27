package ports

import "brokerx/models"

type ComplianceService interface {
	VerifyOrderCompliance(order *models.Order) error
}