package ports

import "brokerx/portfolio-service/models"

type PositionRepository interface {
	FindByUserIdAndSymbol(userId int, symbol string) ([]*models.Position, error)
}
