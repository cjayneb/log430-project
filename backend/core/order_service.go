package core

import (
	"brokerx/models"
	"brokerx/ports"
)

type OrderService struct {
	Repo ports.OrderRepository
}

func (service * OrderService) PlaceOrder(order *models.Order) error {
	_, err := service.Repo.CreateOrder(order)
	if err != nil {
		return err
	}
	return nil
}

var _ ports.OrderService = (*OrderService)(nil) // Ensure interface is implemented at compile time