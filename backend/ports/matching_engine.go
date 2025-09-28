package ports

import "brokerx/models"

type MatchingEngine interface {
	SubmitOrder(order *models.Order) ([]*models.Order, error) // returns orders that match to fill the order passed in args, including the filled order
}