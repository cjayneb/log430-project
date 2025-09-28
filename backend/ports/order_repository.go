package ports

import "brokerx/models"

type OrderRepository interface {
	CreateOrder(order *models.Order) (int, error)
	Update(order *models.Order) error
	FindByUserId(userId string) ([]*models.Order, error)
	FindMatchesMarket(order *models.Order) ([]*models.Order, error)
	FindMatchesLimit(order *models.Order, price float64) ([]*models.Order, error)
}