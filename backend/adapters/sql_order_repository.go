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

func (repo *SQLOrderRepository) FindMatchesMarket(order *models.Order) ([]*models.Order, error) {
    priceOrdering := "ASC"
    if order.Action == "sell" {priceOrdering = "DESC"}
    query := fmt.Sprintf(`
    SELECT id, symbol, action, timing, status, unit_price, remaining_quantity, quantity
        FROM orders
        WHERE symbol = ? AND action <> ? AND status IN ('open','partially_filled')
        ORDER BY unit_price %s, created_at ASC
        FOR UPDATE`, priceOrdering)
    rows, err := repo.tx.QueryContext(context.Background(), query, order.Symbol, order.Action)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var result []*models.Order
    for rows.Next() {
        var o models.Order
        if err := rows.Scan(&o.ID, &o.Symbol, &o.Action, &o.Timing, &o.Status, &o.UnitPrice, &o.RemainingQuantity, &o.Quantity); err != nil {
            return nil, err
        }
        result = append(result, &o)
    }
    return result, nil
}

func (repo *SQLOrderRepository) FindMatchesLimit(order *models.Order, price float64) ([]*models.Order, error) {
    priceOrdering := "<="
    if order.Action == "sell" {priceOrdering = ">="}
    query := fmt.Sprintf(`
    SELECT id, symbol, action, timing, status, unit_price, remaining_quantity, quantity
        FROM orders
        WHERE symbol = ? AND action <> ? AND status IN ('open','partially_filled') AND unit_price %s ?
        ORDER BY created_at ASC
        FOR UPDATE`, priceOrdering)
    rows, err := repo.tx.QueryContext(context.Background(), query, order.Symbol, order.Action, price)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var result []*models.Order
    for rows.Next() {
        var o models.Order
        if err := rows.Scan(&o.ID, &o.Symbol, &o.Action, &o.Timing, &o.Status, &o.UnitPrice, &o.RemainingQuantity, &o.Quantity); err != nil {
            return nil, err
        }
        result = append(result, &o)
    }
    return result, nil
}

func (repo *SQLOrderRepository) Update(order *models.Order) error {
    log.Printf("updating order #%d : %v", order.ID, order)
    _, err := repo.tx.ExecContext(context.Background(),
        `UPDATE orders SET remaining_quantity=?, status=?, unit_price=? WHERE id=?`,
        order.RemainingQuantity, order.Status, order.UnitPrice, order.ID,
    )
    return err
}

func (repo * SQLOrderRepository) CreateOrder(order *models.Order) (int, error) {
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
	rows, err := repo.DB.Query("SELECT symbol, type, action, quantity, unit_price, timing, status FROM brokerx.orders WHERE user_id=?", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order

	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.Symbol, &order.Type, &order.Action, &order.Quantity, &order.UnitPrice, &order.Timing, &order.Status); err != nil {
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