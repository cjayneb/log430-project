package adapters

import (
	"brokerx/models"
	"brokerx/ports"
	"database/sql"
)

type SQLWalletRepository struct {
	DB *sql.DB
}

func (repo *SQLWalletRepository) FindByUserId(userId string) (*models.Wallet, error) {
	row := repo.DB.QueryRow("SELECT available_funds, funds_on_hold FROM brokerx.wallets WHERE user_id=?", userId)

	var wallet models.Wallet
	e := row.Scan(&wallet.AvailableFunds, &wallet.OnHoldFunds)
	if e != nil {
		return nil, e
	}

	return &wallet, nil
}

var _ ports.WalletRepository = (*SQLWalletRepository)(nil) // Ensure interface is implemented at compile time