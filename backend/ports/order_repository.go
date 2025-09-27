package ports

import "brokerx/models"

type OrderRepository interface {
	CreateOrder(order *models.Order) (int, error)
	FindByUserId(userId string) ([]*models.Order, error)
}