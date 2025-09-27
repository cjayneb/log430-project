package ports

import "brokerx/models"

type WalletRepository interface {
	FindByUserId(userId string) (*models.Wallet, error)
}