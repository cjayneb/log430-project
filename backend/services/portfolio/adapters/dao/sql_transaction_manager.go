package dao_adapters

import (
	"brokerx/portfolio-service/ports"
	"brokerx/portfolio-service/util"
	"context"
	"database/sql"
)

type SQLTransactionManager struct {
	DB *sql.DB
}

func (manager *SQLTransactionManager) Do(ctx context.Context, fn func(ports.OrderRepository, ports.ExecutionRepository, ports.WalletRepository, ports.PositionRepository, ports.OutboxRepository) error) error {
	log := util.FromContext(ctx)

	tx, err := manager.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		log.Error("error beginning transaction", "error", err)
		return err
	}
	//nolint
	defer tx.Rollback()

	orderRepo := NewOrderRepo(tx)
	executionRepo := NewExecutionRepo(tx)
	walletRepo := NewWalletRepo(tx)
	positionRepo := NewPositionRepo(tx)
	outboxRepo := NewOutboxRepo(tx)

	if err := fn(orderRepo, executionRepo, walletRepo, positionRepo, outboxRepo); err != nil {
		return err
	}
	return tx.Commit()
}

var _ ports.TransactionManager = (*SQLTransactionManager)(nil) // Ensure interface is implemented at compile time
