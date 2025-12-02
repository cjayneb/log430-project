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

func (repo SQLPositionRepository) ReleaseQuantity(ctx context.Context, deltas []*models.ClaimedCandidate) error {
	log := util.FromContext(ctx)

	if len(deltas) == 0 {
		return nil
	}

	// --- 1. Build the IN clause and lock all rows ---
	placeholders := make([]string, len(deltas))
	for i := range placeholders {
		placeholders[i] = "(?, ?)"
	}
	query := `
        SELECT id, user_id, symbol, available_quantity, reserved_quantity
        FROM brokerx.positions
        WHERE (user_id, symbol) IN (` + strings.Join(placeholders, ",") + `)
        FOR UPDATE
    `

	args := make([]any, 0, len(deltas)*2)
	for _, d := range deltas {
		args = append(args, d.Order.UserID, d.Order.Symbol)
	}

	rows, err := repo.tx.QueryContext(ctx, query, args...)
	if err != nil {
		log.Error("failed to lock rows", "error", err)
		return err
	}
	defer rows.Close()

	// --- 2. Load current values into map ---
	type lockedPos struct {
		ID        int
		UserID    int
		Symbol    string
		Available int
		Reserved  int
	}
	locked := make(map[string]lockedPos)

	for rows.Next() {
		var lp lockedPos

		if err := rows.Scan(&lp.ID, &lp.UserID, &lp.Symbol, &lp.Available, &lp.Reserved); err != nil {
			return err
		}

		key := fmt.Sprintf("%d:%s", lp.UserID, lp.Symbol)
		locked[key] = lp
	}

	// --- 3. Validate constraints (simulate update) ---
	for _, d := range deltas {
		key := fmt.Sprintf("%d:%s", d.Order.UserID, d.Order.Symbol)
		lp, ok := locked[key]
		if !ok {
			return fmt.Errorf("position not found for user=%d symbol=%s", d.Order.UserID, d.Order.Symbol)
		}

		if lp.Reserved < d.ClaimedQty {
			return fmt.Errorf(
				"insufficient quantity for user=%d symbol=%s: reserved=%d needed=%d",
				d.Order.UserID, d.Order.Symbol, lp.Reserved, d.ClaimedQty)
		}

		lp.Reserved -= d.ClaimedQty
		locked[key] = lp
	}

	valueStrings := make([]string, 0, len(locked))
	valueArgs := make([]interface{}, 0, len(locked)*5)

	for _, lp := range locked {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs, lp.ID, lp.UserID, lp.Symbol, lp.Available, lp.Reserved)
	}

	query = fmt.Sprintf(`
		INSERT INTO positions (id, user_id, symbol, available_quantity, reserved_quantity)
		VALUES %s
		ON DUPLICATE KEY UPDATE
			reserved_quantity = VALUES(reserved_quantity);
	`, strings.Join(valueStrings, ","))

	_, err = repo.tx.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		log.Error("Error executing positions batch release", "error", err)
		return err
	}

	return nil
}

func (repo SQLPositionRepository) AddAvailableQuantity(ctx context.Context, deltas []*models.ClaimedCandidate) error {
	log := util.FromContext(ctx)

	if len(deltas) == 0 {
		return nil
	}

	// --- 1. Build the IN clause and lock all rows ---
	placeholders := make([]string, len(deltas))
	for i := range placeholders {
		placeholders[i] = "(?, ?)"
	}
	query := `
        SELECT id, user_id, symbol, available_quantity
        FROM brokerx.positions
        WHERE (user_id, symbol) IN (` + strings.Join(placeholders, ",") + `)
        FOR UPDATE
    `

	args := make([]any, 0, len(deltas)*2)
	for _, d := range deltas {
		args = append(args, d.Order.UserID, d.Order.Symbol)
	}

	rows, err := repo.tx.QueryContext(ctx, query, args...)
	if err != nil {
		log.Error("failed to lock rows", "error", err)
		return err
	}
	defer rows.Close()

	// --- 2. Load current values into map ---
	type lockedPos struct {
		ID        int
		UserID    int
		Symbol    string
		Available int
		Reserved  int
	}
	locked := make(map[string]lockedPos)

	for rows.Next() {
		var lp lockedPos

		if err := rows.Scan(&lp.ID, &lp.UserID, &lp.Symbol, &lp.Available); err != nil {
			return err
		}

		key := fmt.Sprintf("%d:%s", lp.UserID, lp.Symbol)
		locked[key] = lp
	}

	valueStrings := make([]string, 0, len(locked))
	valueArgs := make([]interface{}, 0, len(locked)*4)

	for _, d := range deltas {
		key := fmt.Sprintf("%d:%s", d.Order.UserID, d.Order.Symbol)
		lp, ok := locked[key]
		if !ok {
			lp = lockedPos{UserID: d.Order.UserID, Symbol: d.Order.Symbol}
		}

		lp.Available += d.ClaimedQty
		locked[key] = lp
	}

	for _, lp := range locked {
		valueStrings = append(valueStrings, "(?, ?, ?, ?)")
		valueArgs = append(valueArgs, lp.ID, lp.UserID, lp.Symbol, lp.Available)
	}

	query = fmt.Sprintf(`
		INSERT INTO positions (id, user_id, symbol, available_quantity)
		VALUES %s
		ON DUPLICATE KEY UPDATE
			available_quantity = VALUES(available_quantity);
	`, strings.Join(valueStrings, ","))

	_, err = repo.tx.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		log.Error("Error executing positions batch release", "error", err)
		return err
	}

	return nil
}

func (repo SQLPositionRepository) ReserveQuantity(ctx context.Context, deltas []models.PositionDelta) error {
	log := util.FromContext(ctx)

	if len(deltas) == 0 {
		return nil
	}

	// --- 1. Build the IN clause and lock all rows ---
	placeholders := make([]string, len(deltas))
	for i := range placeholders {
		placeholders[i] = "(?, ?)"
	}
	query := `
        SELECT id, user_id, symbol, available_quantity, reserved_quantity
        FROM brokerx.positions
        WHERE (user_id, symbol) IN (` + strings.Join(placeholders, ",") + `)
        FOR UPDATE
    `

	args := make([]any, 0, len(deltas)*2)
	for _, d := range deltas {
		args = append(args, d.UserID, d.Symbol)
	}

	rows, err := repo.tx.QueryContext(ctx, query, args...)
	if err != nil {
		log.Error("failed to lock rows", "error", err)
		return err
	}
	defer rows.Close()

	// --- 2. Load current values into map ---
	type lockedPos struct {
		ID        int
		UserID    int
		Symbol    string
		Available int
		Reserved  int
	}
	locked := make(map[string]lockedPos)

	for rows.Next() {
		var lp lockedPos

		if err := rows.Scan(&lp.ID, &lp.UserID, &lp.Symbol, &lp.Available, &lp.Reserved); err != nil {
			return err
		}

		key := fmt.Sprintf("%d:%s", lp.UserID, lp.Symbol)
		locked[key] = lp
	}

	// --- 3. Validate constraints (simulate update) ---
	for _, d := range deltas {
		key := fmt.Sprintf("%d:%s", d.UserID, d.Symbol)
		lp, ok := locked[key]
		if !ok {
			return fmt.Errorf("position not found for user=%d symbol=%s", d.UserID, d.Symbol)
		}

		if lp.Available < d.Qty {
			return fmt.Errorf(
				"insufficient quantity for user=%d symbol=%s: available=%d needed=%d",
				d.UserID, d.Symbol, lp.Available, d.Qty)
		}

		lp.Available -= d.Qty
		lp.Reserved += d.Qty
		locked[key] = lp
	}

	valueStrings := make([]string, 0, len(locked))
	valueArgs := make([]interface{}, 0, len(locked)*5)

	for _, lp := range locked {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs, lp.ID, lp.UserID, lp.Symbol, lp.Available, lp.Reserved)
	}

	query = fmt.Sprintf(`
		INSERT INTO positions (id, user_id, symbol, available_quantity, reserved_quantity)
		VALUES %s
		ON DUPLICATE KEY UPDATE
			reserved_quantity = VALUES(reserved_quantity),
			available_quantity = VALUES(available_quantity);
	`, strings.Join(valueStrings, ","))

	_, err = repo.tx.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		log.Error("Error executing positions batch reserve", "error", err)
		return err
	}

	return nil
}

func (repo SQLPositionRepository) RevertReservations(ctx context.Context, deltas []models.PositionDelta) error {
	log := util.FromContext(ctx)

	if len(deltas) == 0 {
		return nil
	}

	// --- 1. Build the IN clause and lock all rows ---
	placeholders := make([]string, len(deltas))
	for i := range placeholders {
		placeholders[i] = "(?, ?)"
	}
	query := `
        SELECT id, user_id, symbol, available_quantity, reserved_quantity
        FROM brokerx.positions
        WHERE (user_id, symbol) IN (` + strings.Join(placeholders, ",") + `)
        FOR UPDATE
    `

	args := make([]any, 0, len(deltas)*2)
	for _, d := range deltas {
		args = append(args, d.UserID, d.Symbol)
	}

	rows, err := repo.tx.QueryContext(ctx, query, args...)
	if err != nil {
		log.Error("failed to lock rows", "error", err)
		return err
	}
	defer rows.Close()

	// --- 2. Load current values into map ---
	locked := make(map[string]models.Position)

	for rows.Next() {
		var lp models.Position

		if err := rows.Scan(&lp.ID, &lp.UserId, &lp.Symbol, &lp.AvailableQuantity, &lp.ReservedQuantity); err != nil {
			return err
		}

		key := fmt.Sprintf("%d:%s", lp.UserId, lp.Symbol)
		locked[key] = lp
	}

	// --- 3. Validate constraints (simulate update) ---
	for _, d := range deltas {
		key := fmt.Sprintf("%d:%s", d.UserID, d.Symbol)
		lp, ok := locked[key]
		if !ok {
			return fmt.Errorf("position not found for user=%d symbol=%s", d.UserID, d.Symbol)
		}

		if lp.ReservedQuantity < d.Qty {
			return fmt.Errorf(
				"insufficient quantity to revert for user=%d symbol=%s: reserved=%d needed=%d",
				d.UserID, d.Symbol, lp.AvailableQuantity, d.Qty)
		}

		lp.AvailableQuantity += d.Qty
		lp.ReservedQuantity -= d.Qty
		locked[key] = lp
	}

	valueStrings := make([]string, 0, len(locked))
	valueArgs := make([]interface{}, 0, len(locked)*5)

	for _, lp := range locked {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs, lp.ID, lp.UserId, lp.Symbol, lp.AvailableQuantity, lp.ReservedQuantity)
	}

	query = fmt.Sprintf(`
		INSERT INTO positions (id, user_id, symbol, available_quantity, reserved_quantity)
		VALUES %s
		ON DUPLICATE KEY UPDATE
			reserved_quantity = VALUES(reserved_quantity),
			available_quantity = VALUES(available_quantity);
	`, strings.Join(valueStrings, ","))

	_, err = repo.tx.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		log.Error("Error executing positions batch revert reservation", "error", err)
		return err
	}

	return nil
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

	rows, err := repo.DB.Query("SELECT symbol, available_quantity, reserved_quantity, unit_price FROM brokerx.positions WHERE user_id=? and symbol=?", userId, symbol)
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
		if err := rows.Scan(&pos.Symbol, &pos.AvailableQuantity, &pos.ReservedQuantity, &pos.UnitPrice); err != nil {
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
