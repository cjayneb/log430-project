package core

import (
	"brokerx/models"
	"brokerx/ports"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var orderQueueLen = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "order_queue_length",
	Help: "Current number of orders in the matching queue",
})

func init() { prometheus.MustRegister(orderQueueLen) }

var orderQueue chan *models.Order

func StartMatchingWorkers(matchingEngine ports.MatchingEngine) {
	orderQueue = make(chan *models.Order, 1000)
	for i := 0; i < 8; i++ {
		go func() {
			for order := range orderQueue {
				if err := matchingEngine.SubmitOrder(order); err != nil {
					log.Errorf("matching failed for order #%d: %v", order.ID, err)
				}
			}
		}()
	}
	log.Infof("Started %d matching workers", 8)
}

type OrderService struct {
	Repo              ports.OrderRepository
	ComplianceService ports.ComplianceService
	MatchingEngine    ports.MatchingEngine
}

func (service *OrderService) PlaceOrder(order *models.Order) error {
	err := service.ComplianceService.VerifyOrderCompliance(order)
	if err != nil {
		log.Errorf("Error when verifying order compliance : %v", err)
		return err
	}

	createdOrderId, err := service.Repo.CreateOrder(order)
	if err != nil {
		return err
	}
	order.ID = createdOrderId

	orderQueueLen.Set(float64(len(orderQueue)))
	orderQueue <- order
	log.Infof("Order #%v queued for matching", order.ID)
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
