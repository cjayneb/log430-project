package adapters

import (
	"brokerx/models"
	"brokerx/ports"
	"context"
	"database/sql"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type SQLOrderRepository struct {
	DB *sql.DB
	tx *sql.Tx
}

func NewOrderRepo(tx *sql.Tx) ports.OrderRepository {
	return &SQLOrderRepository{tx: tx}
}

func (repo *SQLOrderRepository) FindMatchesMarket(order *models.Order, limit int) ([]*models.Order, error) {
	priceOrdering := "ASC"
	if order.Action == "sell" {
		priceOrdering = "DESC"
	}
	query := fmt.Sprintf(`
    SELECT id, user_id, symbol, action, type, timing, status, unit_price, remaining_quantity, quantity
        FROM orders
        WHERE symbol = ? AND action <> ? AND status IN ('open','partially_filled') AND type <> 'market'
        ORDER BY unit_price %s, created_at ASC
        LIMIT ?`, priceOrdering)
	rows, err := repo.tx.QueryContext(context.Background(), query, order.Symbol, order.Action, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Symbol, &o.Action, &o.Type, &o.Timing, &o.Status, &o.UnitPrice, &o.RemainingQuantity, &o.Quantity); err != nil {
			return nil, err
		}
		result = append(result, &o)
	}
	return result, nil
}

func (repo *SQLOrderRepository) FindMatchesLimit(order *models.Order, price float64, limit int) ([]*models.Order, error) {
	priceComparison := "<="
	priceOrdering := "ASC"
	if order.Action == "sell" {
		priceComparison = ">="
		priceOrdering = "DESC"
	}
	query := fmt.Sprintf(`
    SELECT id, user_id, symbol, action, type, timing, status, unit_price, remaining_quantity, quantity
        FROM orders
        WHERE symbol = ? AND action <> ? AND status IN ('open','partially_filled') AND (type = 'market' OR unit_price %s ?)
        ORDER BY type ASC, unit_price %s, created_at ASC
        LIMIT ?`, priceComparison, priceOrdering)
	rows, err := repo.tx.QueryContext(context.Background(), query, order.Symbol, order.Action, price, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Symbol, &o.Action, &o.Type, &o.Timing, &o.Status, &o.UnitPrice, &o.RemainingQuantity, &o.Quantity); err != nil {
			return nil, err
		}
		result = append(result, &o)
	}
	return result, nil
}

func (repo *SQLOrderRepository) ClaimOrder(orderID int, unitPrice float64, qty int) (int64, error) {
	res, err := repo.tx.ExecContext(context.Background(), `
		UPDATE orders
		SET remaining_quantity = remaining_quantity - ?,
			status = CASE
				WHEN remaining_quantity - ? = 0 THEN 'filled'
				WHEN remaining_quantity - ? > 0 THEN 'partially_filled'
				ELSE status END,
			unit_price = ?
		WHERE id = ? AND remaining_quantity >= ? AND status IN ('open','partially_filled');`,
		qty, qty, qty, unitPrice, orderID, qty)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (repo *SQLOrderRepository) RevertClaim(orderID int, unitPrice float64, qty int) error {
	_, err := repo.tx.ExecContext(context.Background(), `
        UPDATE orders
        SET remaining_quantity = remaining_quantity + ?,
            status = CASE
                WHEN remaining_quantity + ? = quantity THEN 'open'
                ELSE 'partially_filled' END
			unit_price = ?
        WHERE id = ?;`, qty, qty, unitPrice, orderID)
	return err
}

func (repo *SQLOrderRepository) Update(order *models.Order) error {
	_, err := repo.tx.ExecContext(context.Background(),
		`UPDATE orders SET remaining_quantity=?, status=?, unit_price=? WHERE id=?`,
		order.RemainingQuantity, order.Status, order.UnitPrice, order.ID,
	)
	return err
}

func (repo *SQLOrderRepository) CreateOrder(order *models.Order) (int, error) {
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
	rows, err := repo.DB.Query("SELECT id, symbol, type, action, quantity, remaining_quantity, unit_price, timing, status, created_at FROM brokerx.orders WHERE user_id=? ORDER BY created_at DESC LIMIT 100", userId)
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
