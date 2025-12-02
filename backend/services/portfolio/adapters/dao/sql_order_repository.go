package dao_adapters

import (
	"brokerx/portfolio-service/models"
	"brokerx/portfolio-service/ports"
	"brokerx/portfolio-service/util"
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
	log := util.FromContext(ctx)

	if len(orders) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(orders))
	valueArgs := make([]interface{}, 0, len(orders)*9)

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

var _ ports.OrderRepository = (*SQLOrderRepository)(nil) // Ensure interface is implemented at compile time
