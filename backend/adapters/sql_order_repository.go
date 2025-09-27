package adapters

import (
	"brokerx/models"
	"brokerx/ports"
	"database/sql"

	log "github.com/sirupsen/logrus"
)

type SQLOrderRepository struct {
	DB *sql.DB
}

func (repo * SQLOrderRepository) CreateOrder(order *models.Order) (int, error) {
	result, err := repo.DB.Exec("INSERT INTO orders (user_id, symbol, type, action, quantity, unit_price, timing, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		order.UserID, order.Symbol, order.Type, order.Action, order.Quantity, order.UnitPrice, order.Timing, order.Status)
	if err != nil {
		log.Errorf("Error creating order: %v", err)
		return 0, err
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

var _ ports.OrderRepository = (*SQLOrderRepository)(nil) // Ensure interface is implemented at compile time