package ports

import "brokerx/models"

type OrderBook interface {
	GetById(orderId int) (models.Order, error)
	FetchByIDs(ids []string) ([]*models.Order, error)
	FindMatchesLimit(symbol string, orderType string, action string, unitPrice float64, batchSize int) ([]*models.Order, error)
	FindMatchesMarket(symbol string, orderType string, action string, batchSize int) ([]*models.Order, error)
	Insert(order *models.Order) error
	Return(orders []*models.Order) error
}
