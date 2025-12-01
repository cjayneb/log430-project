package dao_adapters

import (
	"brokerx/order-service/common"
	"brokerx/order-service/models"
	"brokerx/order-service/ports"
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type SQLOrderRepository struct {
	DB *sql.DB
	tx *sql.Tx
}

func NewOrderRepo(tx *sql.Tx) SQLOrderRepository {
	return SQLOrderRepository{tx: tx}
}

func (repo SQLOrderRepository) UpdateBatch(ctx context.Context, orders []*models.Order) error {
	log := common.FromContext(ctx)

	if len(orders) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(orders))
	valueArgs := make([]interface{}, 0, len(orders)*5)

	for _, o := range orders {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs, o.ID, o.UserID, o.Symbol, o.Type, o.Action, o.RemainingQuantity, o.Status, o.UnitPrice, o.Timing)
	}

	query := fmt.Sprintf(`
		INSERT INTO orders (id, user_id, symbol, type, action, remaining_quantity, status, unit_price, timing)
		VALUES %s
		ON DUPLICATE KEY UPDATE
			remaining_quantity = VALUES(remaining_quantity),
			status = VALUES(status),
			unit_price = VALUES(unit_price);
	`, strings.Join(valueStrings, ","))

	_, err := repo.tx.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		log.Error("Error executing batch upsert", "error", err)
	}
	return err
}

func (repo SQLOrderRepository) Create(ctx context.Context, order *models.Order) (int, error) {
	log := common.FromContext(ctx)

	result, err := repo.DB.Exec("INSERT INTO orders (user_id, symbol, type, action, quantity, remaining_quantity, unit_price, timing, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		order.UserID, order.Symbol, order.Type, order.Action, order.Quantity, order.RemainingQuantity, order.UnitPrice, order.Timing, order.Status)
	if err != nil {
		log.Error("Error creating order", "error", err)
		return 0, err
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

func (repo SQLOrderRepository) FindByUserId(ctx context.Context, userId int) ([]*models.Order, error) {
	log := common.FromContext(ctx)

	rows, err := repo.DB.Query("SELECT id, symbol, type, action, quantity, remaining_quantity, unit_price, timing, status, created_at FROM brokerx.orders WHERE user_id=? ORDER BY created_at DESC, id DESC LIMIT 100", userId)
	if err != nil {
		log.Error("error when fetching orders", "error", err)
		return nil, err
	}
	defer rows.Close()

	orders := []*models.Order{}

	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.ID, &order.Symbol, &order.Type, &order.Action, &order.Quantity, &order.RemainingQuantity, &order.UnitPrice, &order.Timing, &order.Status, &order.CreatedAt); err != nil {
			log.Error("error when reading fetched orders", "error", err)
			return nil, err
		}
		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		log.Error("row erros", "error", err)
		return nil, err
	}

	return orders, nil
}

var _ ports.OrderRepository = (*SQLOrderRepository)(nil) // Ensure interface is implemented at compile time
