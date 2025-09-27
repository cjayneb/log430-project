package ports

import "brokerx/models"

type OrderService interface {
    PlaceOrder(order *models.Order) error
}
