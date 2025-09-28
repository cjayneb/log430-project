package ports

import "brokerx/models"

type MatchingEngine interface {
	SubmitOrder(order *models.Order) error
}