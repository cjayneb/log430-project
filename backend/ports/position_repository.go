package ports

import "brokerx/models"

type PositionRepository interface {
	FindByUserIdAndSymbol(userId string, symbol string) ([]*models.Position, error)
}