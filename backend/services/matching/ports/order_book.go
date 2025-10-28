package ports

import "brokerx/matching-service/models"

type OrderBook interface {
	GetById(orderId int) (models.Order, error)
	FindMatchesLimit(symbol string, action string, unitPrice float64, batchSize int) ([]*models.Order, error)
	FindMatchesMarket(symbol string, action string, batchSize int) ([]*models.Order, error)
	Insert(order *models.Order) error
	Return(orders []*models.Order) error
	EnqueueOrders(orders []*models.Order) error
	DequeueOrders(batchSize int) ([]*models.Order, error)
}
