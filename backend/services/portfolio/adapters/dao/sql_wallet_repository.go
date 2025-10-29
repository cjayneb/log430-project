package dao_adapters

import (
	"brokerx/portfolio-service/models"
	"brokerx/portfolio-service/ports"
	"database/sql"
	"errors"
	"fmt"
)

type SQLWalletRepository struct {
	DB *sql.DB
}

func (repo *SQLWalletRepository) AddFunds(userId int, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	tx, err := repo.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	
	res, err := tx.Exec(`
		UPDATE wallets 
		SET available_funds = available_funds + ? 
		WHERE user_id = ?`,
		amount, userId)
	if err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	// If no wallet exists, insert one
	if rowsAffected == 0 {
		_, err = tx.Exec(`
			INSERT INTO wallets (id, user_id, available_funds, funds_on_hold)
			VALUES (UUID(), ?, ?, 0)`,
			userId, amount)
		if err != nil {
			return fmt.Errorf("failed to create wallet: %w", err)
		}
	}

	return nil
}

func (repo *SQLWalletRepository) FindByUserId(userId int) (*models.Wallet, error) {
	row := repo.DB.QueryRow("SELECT available_funds, funds_on_hold FROM brokerx.wallets WHERE user_id=?", userId)

	var wallet models.Wallet
	e := row.Scan(&wallet.AvailableFunds, &wallet.OnHoldFunds)
	if e == sql.ErrNoRows {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}

	return &wallet, nil
}

var _ ports.WalletRepository = (*SQLWalletRepository)(nil) // Ensure interface is implemented at compile time
