package dao_adapters

import (
	"brokerx/portfolio-service/models"
	"brokerx/portfolio-service/ports"
	"brokerx/portfolio-service/util"
	"context"
	"database/sql"
)

type SQLPositionRepository struct {
	DB *sql.DB
	tx *sql.Tx
}

func NewPositionRepo(tx *sql.Tx) SQLPositionRepository {
	return SQLPositionRepository{tx: tx}
}

// Update implements ports.PositionRepository.
func (repo SQLPositionRepository) Update(ctx context.Context, userId int, symbol string, qty int) error {
	panic("unimplemented")
}

func (repo SQLPositionRepository) FindByUserIdAndSymbol(ctx context.Context, userId int, symbol string) ([]*models.Position, error) {
	log := util.FromContext(ctx)

	rows, err := repo.DB.Query("SELECT symbol, quantity, unit_price FROM brokerx.positions WHERE user_id=? and symbol=?", userId, symbol)
	if err == sql.ErrNoRows {
		return []*models.Position{}, nil
	}
	if err != nil {
		log.Error("error executing query", "error", err)
		return nil, err
	}
	defer rows.Close()

	positions := []*models.Position{}

	for rows.Next() {
		var pos models.Position
		if err := rows.Scan(&pos.Symbol, &pos.Quantity, &pos.UnitPrice); err != nil {
			log.Error("error scanning row", "error", err)
			return nil, err
		}
		positions = append(positions, &pos)
	}

	if err := rows.Err(); err != nil {
		log.Error("error found in rows", "error", err)
		return nil, err
	}

	return positions, nil
}

var _ ports.PositionRepository = (*SQLPositionRepository)(nil) // Ensure interface is implemented at compile time
