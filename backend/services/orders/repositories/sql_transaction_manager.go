package repositories

import (
	"brokerx/order-service/ports"
	"context"
	"database/sql"
)

type SQLTransactionManager struct {
	DB *sql.DB
}

func (manager *SQLTransactionManager) Do(ctx context.Context, fn func(ports.OrderRepository, ports.ExecutionRepository) error) error {
	tx, err := manager.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	//nolint
	defer tx.Rollback()

	orderRepo := NewOrderRepo(tx)
	executionRepo := NewExecutionRepo(tx)

	if err := fn(orderRepo, executionRepo); err != nil {
		return err
	}
	return tx.Commit()
}

var _ ports.TransactionManager = (*SQLTransactionManager)(nil) // Ensure interface is implemented at compile time
