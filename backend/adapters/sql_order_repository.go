package adapters

import (
	"brokerx/models"
	"brokerx/ports"
	"context"
	"database/sql"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

type SQLOrderRepository struct {
	DB *sql.DB
	tx *sql.Tx
}

func NewOrderRepo(tx *sql.Tx) ports.OrderRepository {
	return &SQLOrderRepository{tx: tx}
}

func (repo *SQLOrderRepository) Update(order *models.Order) error {
	_, err := repo.tx.ExecContext(context.Background(),
		`UPDATE orders SET remaining_quantity=?, status=?, unit_price=? WHERE id=?`,
		order.RemainingQuantity, order.Status, order.UnitPrice, order.ID,
	)
	return err
}

func (repo *SQLOrderRepository) UpdateBatch(orders []*models.Order) error {
	if len(orders) == 0 {
		return nil
	}

	ctx := context.Background()

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
		log.Errorf("Error executing batch upsert: %v", err)
	}
	return err
}


func (repo *SQLOrderRepository) Create(order *models.Order) (int, error) {
	result, err := repo.DB.Exec("INSERT INTO orders (user_id, symbol, type, action, quantity, remaining_quantity, unit_price, timing, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		order.UserID, order.Symbol, order.Type, order.Action, order.Quantity, order.RemainingQuantity, order.UnitPrice, order.Timing, order.Status)
	if err != nil {
		log.Errorf("Error creating order: %v", err)
		return 0, err
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

func (repo *SQLOrderRepository) FindByUserId(userId string) ([]*models.Order, error) {
	rows, err := repo.DB.Query("SELECT id, symbol, type, action, quantity, remaining_quantity, unit_price, timing, status, created_at FROM brokerx.orders WHERE user_id=? ORDER BY created_at DESC, id DESC LIMIT 100", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order

	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.ID, &order.Symbol, &order.Type, &order.Action, &order.Quantity, &order.RemainingQuantity, &order.UnitPrice, &order.Timing, &order.Status, &order.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

var _ ports.OrderRepository = (*SQLOrderRepository)(nil) // Ensure interface is implemented at compile time
