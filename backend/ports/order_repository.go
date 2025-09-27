package ports

import "brokerx/models"

type OrderRepository interface {
	CreateOrder(order *models.Order) (int, error)
}