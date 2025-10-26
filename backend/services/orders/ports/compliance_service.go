package ports

import "brokerx/order-service/models"

type ComplianceService interface {
	VerifyOrderCompliance(order *models.Order) error
}
