package ports

import "brokerx/order-service/models"

type OrderRepository interface {
	Create(order *models.Order) (int, error)
	Update(order *models.Order) error
	UpdateBatch(orders []*models.Order) error
	FindByUserId(userId string) ([]*models.Order, error)
}
