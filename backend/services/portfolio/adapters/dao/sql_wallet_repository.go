package dao_adapters

import (
	"brokerx/portfolio-service/models"
	"brokerx/portfolio-service/ports"
	"brokerx/portfolio-service/util"
	"context"
	"database/sql"
)

type SQLWalletRepository struct {
	DB *sql.DB
}

func (repo *SQLWalletRepository) AddFunds(ctx context.Context, userId int, amount float64) error {
	log := util.FromContext(ctx)

	tx, err := repo.DB.Begin()
	if err != nil {
		log.Error("error starting transaction", "error", err)
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			log.Warn("transaction failed, rolling back")
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	
	_, err = tx.Exec(`
		UPDATE wallets 
		SET available_funds = available_funds + ? 
		WHERE user_id = ?`,
		amount, userId)
	if err != nil {
		log.Error("error updating wallet", "error", err)
		return err
	}

	return nil
}

func (repo *SQLWalletRepository) FindByUserId(ctx context.Context, userId int) (*models.Wallet, error) {
	log := util.FromContext(ctx)

	row := repo.DB.QueryRow("SELECT available_funds, funds_on_hold FROM brokerx.wallets WHERE user_id=?", userId)

	var wallet models.Wallet
	e := row.Scan(&wallet.AvailableFunds, &wallet.OnHoldFunds)
	if e == sql.ErrNoRows {
		return nil, nil
	}
	if e != nil {
		log.Error("error fetching wallet", "error", e)
		return nil, e
	}

	return &wallet, nil
}

var _ ports.WalletRepository = (*SQLWalletRepository)(nil) // Ensure interface is implemented at compile time
