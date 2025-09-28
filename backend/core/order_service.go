package core

import (
	"brokerx/models"
	"brokerx/ports"

	log "github.com/sirupsen/logrus"
)

type OrderService struct {
	Repo ports.OrderRepository
	ComplianceService ports.ComplianceService
	MatchingEngine ports.MatchingEngine
}

func (service * OrderService) PlaceOrder(order *models.Order) error {
	err := service.ComplianceService.VerifyOrderCompliance(order)
	if err != nil {
		log.Errorf("Error when verifying order compliance : %v", err)
		return err
	}

	_, err = service.Repo.CreateOrder(order)
	if err != nil {
		return err
	}

	matchedOrders, err := service.MatchingEngine.SubmitOrder(order)
	if err != nil {
		return err
	}

	if len(matchedOrders) > 0 {
		service.Repo.Update(matchedOrders)
	}

	return nil
}

func (service *OrderService) GetOrdersForUser(userId string) ([]*models.Order, error) {
	orders, err := service.Repo.FindByUserId(userId)
	if err != nil {
		log.Errorf("Error when fetching user orders : %v", err)
		return nil, err
	}

	return orders, nil
}

var _ ports.OrderService = (*OrderService)(nil) // Ensure interface is implemented at compile time