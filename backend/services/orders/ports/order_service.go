package ports

import "brokerx/order-service/models"

type OrderService interface {
	PlaceOrder(order *models.Order) error
	GetOrdersForUser(userId string) ([]*models.Order, error)
}
