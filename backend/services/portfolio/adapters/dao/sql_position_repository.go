package dao_adapters

import (
	"brokerx/portfolio-service/models"
	"brokerx/portfolio-service/ports"
	"brokerx/portfolio-service/util"
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type SQLPositionRepository struct {
	DB *sql.DB
	tx *sql.Tx
}

func NewPositionRepo(tx *sql.Tx) SQLPositionRepository {
	return SQLPositionRepository{tx: tx}
}

func posKey(userID int, symbol string) string {
	return fmt.Sprintf("%d:%s", userID, symbol)
}

func (repo SQLPositionRepository) lockPositions(
	ctx context.Context,
	keys [][2]any,
	queryFields string,
	scanFn func(*sql.Rows) (models.Position, error),
) (map[string]models.Position, error) {
	log := util.FromContext(ctx)

	placeholders := make([]string, len(keys))
	for i := range placeholders {
		placeholders[i] = "(?, ?)"
	}

	query := fmt.Sprintf(`
        SELECT %s
        FROM brokerx.positions
        WHERE (user_id, symbol) IN (%s)
        FOR UPDATE
    `, queryFields, strings.Join(placeholders, ","))

	args := make([]any, 0, len(keys)*2)
	for _, k := range keys {
		args = append(args, k[0], k[1])
	}

	rows, err := repo.tx.QueryContext(ctx, query, args...)
	if err != nil {
		log.Error("failed to lock rows", "error", err)
		return nil, err
	}
	defer rows.Close()

	locked := make(map[string]models.Position)
	for rows.Next() {
		lp, err := scanFn(rows)
		if err != nil {
			log.Error("error scanning rows", "error", err)
			return nil, err
		}
		locked[posKey(lp.UserId, lp.Symbol)] = lp
	}

	return locked, nil
}

func (repo SQLPositionRepository) upsertPositions(
	ctx context.Context,
	locked map[string]models.Position,
	fields string,
	update string,
) error {
	valueStrings := make([]string, 0, len(locked))
	valueArgs := make([]interface{}, 0, len(locked)*5)



	for _, lp := range locked {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs, lp.ID, lp.UserId, lp.Symbol, lp.AvailableQuantity, lp.ReservedQuantity)
	}

	query := fmt.Sprintf(`
		INSERT INTO positions %s
		VALUES %s
		ON DUPLICATE KEY UPDATE %s
	`, fields, strings.Join(valueStrings, ","), update)

	_, err := repo.tx.ExecContext(ctx, query, valueArgs...)
	return err
}



func (repo SQLPositionRepository) ReleaseQuantity(ctx context.Context, deltas []*models.ClaimedCandidate) error {
	log := util.FromContext(ctx)
	if len(deltas) == 0 {
		return nil
	}

	keys := make([][2]any, 0, len(deltas))
	for _, d := range deltas {
		keys = append(keys, [2]any{d.Order.UserID, d.Order.Symbol})
	}

	locked, err := repo.lockPositions(ctx, keys,
		"id, user_id, symbol, available_quantity, reserved_quantity",
		func(r *sql.Rows) (models.Position, error) {
			var lp models.Position
			err := r.Scan(&lp.ID, &lp.UserId, &lp.Symbol, &lp.AvailableQuantity, &lp.ReservedQuantity)
			return lp, err
		})
	if err != nil {
		return err
	}

	for _, d := range deltas {
		key := posKey(d.Order.UserID, d.Order.Symbol)
		lp, ok := locked[key]
		if !ok {
			return fmt.Errorf("position not found for user=%d symbol=%s", d.Order.UserID, d.Order.Symbol)
		}
		if lp.ReservedQuantity < d.ClaimedQty {
			return fmt.Errorf("insufficient quantity: reserved=%d need=%d", lp.ReservedQuantity, d.ClaimedQty)
		}
		lp.ReservedQuantity -= d.ClaimedQty
		locked[key] = lp
	}

	err = repo.upsertPositions(
		ctx,
		locked,
		"(id, user_id, symbol, available_quantity, reserved_quantity)",
		"reserved_quantity = VALUES(reserved_quantity)",
	)
	if err != nil {
		log.Error("error updating reserved quantities", "error", err)
	}
	return err
}

func (repo SQLPositionRepository) AddAvailableQuantity(ctx context.Context, deltas []*models.ClaimedCandidate) error {
	log := util.FromContext(ctx)
	if len(deltas) == 0 {
		return nil
	}

	keys := make([][2]any, 0, len(deltas))
	for _, d := range deltas {
		keys = append(keys, [2]any{d.Order.UserID, d.Order.Symbol})
	}

	locked, err := repo.lockPositions(ctx, keys,
		"id, user_id, symbol, available_quantity, reserved_quantity",
		func(r *sql.Rows) (models.Position, error) {
			var lp models.Position
			err := r.Scan(&lp.ID, &lp.UserId, &lp.Symbol, &lp.AvailableQuantity, &lp.ReservedQuantity)
			return lp, err
		})
	if err != nil {
		return err
	}

	for _, d := range deltas {
		key := posKey(d.Order.UserID, d.Order.Symbol)
		lp := locked[key]
		lp.UserId = d.Order.UserID
		lp.Symbol = d.Order.Symbol
		lp.AvailableQuantity += d.ClaimedQty
		locked[key] = lp
	}

	err = repo.upsertPositions(
		ctx,
		locked,
		"(id, user_id, symbol, available_quantity, reserved_quantity)",
		"available_quantity = VALUES(available_quantity)",
	)
	if err != nil {
		log.Error("error updating available quantities", "error", err)
	}
	return err
}

func (repo SQLPositionRepository) ReserveQuantity(ctx context.Context, deltas []models.PositionDelta) error {
	log := util.FromContext(ctx)
	if len(deltas) == 0 {
		return nil
	}

	keys := make([][2]any, 0, len(deltas))
	for _, d := range deltas {
		keys = append(keys, [2]any{d.UserID, d.Symbol})
	}

	locked, err := repo.lockPositions(ctx, keys,
		"id, user_id, symbol, available_quantity, reserved_quantity",
		func(r *sql.Rows) (models.Position, error) {
			var lp models.Position
			err := r.Scan(&lp.ID, &lp.UserId, &lp.Symbol, &lp.AvailableQuantity, &lp.ReservedQuantity)
			return lp, err
		})
	if err != nil {
		return err
	}

	for _, d := range deltas {
		key := posKey(d.UserID, d.Symbol)
		lp, ok := locked[key]
		if !ok {
			return fmt.Errorf("position not found for user=%d symbol=%s", d.UserID, d.Symbol)
		}
		if lp.AvailableQuantity < d.Qty {
			return fmt.Errorf("insufficient quantity available=%d need=%d", lp.AvailableQuantity, d.Qty)
		}
		lp.AvailableQuantity -= d.Qty
		lp.ReservedQuantity += d.Qty
		locked[key] = lp
	}

	err = repo.upsertPositions(
		ctx,
		locked,
		"(id, user_id, symbol, available_quantity, reserved_quantity)",
		"reserved_quantity = VALUES(reserved_quantity), available_quantity = VALUES(available_quantity)",
	)
	if err != nil {
		log.Error("error updating reserved quantities", "error", err)
	}
	return err
}

func (repo SQLPositionRepository) RevertReservations(ctx context.Context, deltas []models.PositionDelta) error {
	log := util.FromContext(ctx)
	if len(deltas) == 0 {
		return nil
	}

	keys := make([][2]any, 0, len(deltas))
	for _, d := range deltas {
		keys = append(keys, [2]any{d.UserID, d.Symbol})
	}

	locked, err := repo.lockPositions(ctx, keys,
		"id, user_id, symbol, available_quantity, reserved_quantity",
		func(r *sql.Rows) (models.Position, error) {
			var lp models.Position
			err := r.Scan(&lp.ID, &lp.UserId, &lp.Symbol, &lp.AvailableQuantity, &lp.ReservedQuantity)
			return lp, err
		})
	if err != nil {
		return err
	}

	for _, d := range deltas {
		key := posKey(d.UserID, d.Symbol)
		lp, ok := locked[key]
		if !ok {
			return fmt.Errorf("position not found for user=%d symbol=%s", d.UserID, d.Symbol)
		}
		if lp.ReservedQuantity < d.Qty {
			return fmt.Errorf("insufficient reserved quantity=%d need=%d", lp.ReservedQuantity, d.Qty)
		}
		lp.ReservedQuantity -= d.Qty
		lp.AvailableQuantity += d.Qty
		locked[key] = lp
	}

	err = repo.upsertPositions(
		ctx,
		locked,
		"(id, user_id, symbol, available_quantity, reserved_quantity)",
		"reserved_quantity = VALUES(reserved_quantity), available_quantity = VALUES(available_quantity)",
	)
	if err != nil {
		log.Error("error updating reserved quantities", "error", err)
	}
	return err
}

func (repo SQLPositionRepository) FindByUserId(ctx context.Context, userId int) ([]*models.Position, error) {
	log := util.FromContext(ctx)

	rows, err := repo.DB.Query("SELECT id, symbol, available_quantity, reserved_quantity FROM brokerx.positions WHERE user_id=?", userId)
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
		if err := rows.Scan(&pos.ID, &pos.Symbol, &pos.AvailableQuantity, &pos.ReservedQuantity); err != nil {
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

func (repo SQLPositionRepository) FindByUserIdAndSymbol(ctx context.Context, userId int, symbol string) ([]*models.Position, error) {
	log := util.FromContext(ctx)

	rows, err := repo.DB.Query("SELECT symbol, available_quantity, reserved_quantity FROM brokerx.positions WHERE user_id=? and symbol=?", userId, symbol)
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
		if err := rows.Scan(&pos.Symbol, &pos.AvailableQuantity, &pos.ReservedQuantity); err != nil {
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
