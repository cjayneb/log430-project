package dao_adapters

import (
	"brokerx/portfolio-service/models"
	"brokerx/portfolio-service/ports"
	"database/sql"
)

type SQLPositionRepository struct {
	DB *sql.DB
}

func (repo *SQLPositionRepository) FindByUserIdAndSymbol(userId int, symbol string) ([]*models.Position, error) {
	rows, err := repo.DB.Query("SELECT symbol, quantity, unit_price FROM brokerx.positions WHERE user_id=? and symbol=?", userId, symbol)
	if err == sql.ErrNoRows {
		return []*models.Position{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	positions := []*models.Position{}

	for rows.Next() {
		var pos models.Position
		if err := rows.Scan(&pos.Symbol, &pos.Quantity, &pos.UnitPrice); err != nil {
			return nil, err
		}
		positions = append(positions, &pos)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return positions, nil
}

var _ ports.PositionRepository = (*SQLPositionRepository)(nil) // Ensure interface is implemented at compile time
