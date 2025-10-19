package ports

import "brokerx/models"

type OrderRepository interface {
	CreateOrder(order *models.Order) (int, error)
	Update(order *models.Order) error
	FindByUserId(userId string) ([]*models.Order, error)
	FindMatchesMarket(order *models.Order, limit int) ([]*models.Order, error)
	FindMatchesLimit(order *models.Order, price float64, limit int) ([]*models.Order, error)
	ClaimOrder(orderID int, unitPrice float64, qty int) (int64, error)
	RevertClaim(orderID int, unitPrice float64, qty int) error
}
