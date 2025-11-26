package dao_adapters

import (
	"brokerx/portfolio-service/models"
	"brokerx/portfolio-service/ports"
	"context"
	"database/sql"
	"strings"
)

type SQLExecutionRepository struct {
	DB *sql.DB
	tx *sql.Tx
}

func NewExecutionRepo(tx *sql.Tx) ports.ExecutionRepository {
	return &SQLExecutionRepository{tx: tx}
}

func (repo *SQLExecutionRepository) CreateBatch(ctx context.Context, execs []*models.ExecutionRecord) error {
	if len(execs) == 0 {
		return nil
	}

	query := "INSERT INTO executions (buy_order_id, sell_order_id, symbol, unit_price, quantity) VALUES "
	args := []interface{}{}
	placeholders := []string{}

	for _, e := range execs {
		placeholders = append(placeholders, "(?, ?, ?, ?, ?)")
		args = append(args, e.BuyOrderID, e.SellOrderID, e.Symbol, e.Price, e.Quantity)
	}

	query += strings.Join(placeholders, ",")
	_, err := repo.DB.ExecContext(context.Background(), query, args...)
	return err
}

var _ ports.ExecutionRepository = (*SQLExecutionRepository)(nil) // Ensure interface is implemented at compile time
