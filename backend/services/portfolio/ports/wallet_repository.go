package ports

import "brokerx/portfolio-service/models"

type WalletRepository interface {
	FindByUserId(userId int) (*models.Wallet, error)
	AddFunds(userId int, amount float64) error
}
