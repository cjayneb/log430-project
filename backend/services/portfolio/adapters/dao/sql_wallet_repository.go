package dao_adapters

import (
	"brokerx/portfolio-service/models"
	"brokerx/portfolio-service/ports"
	"brokerx/portfolio-service/util"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type SQLWalletRepository struct {
	DB *sql.DB
	tx *sql.Tx
}

func NewWalletRepo(tx *sql.Tx) SQLWalletRepository {
	return SQLWalletRepository{tx: tx}
}

func (repo SQLWalletRepository) ReserveFunds(ctx context.Context, userId int, amount float64) error {
	log := util.FromContext(ctx)

	w, err := repo.loadWalletForUpdate(ctx, userId)
	if err != nil {
		return err
	}

	if w.AvailableFunds < amount {
		return fmt.Errorf("insufficient funds: available=%f, required=%f", w.AvailableFunds, amount)
	}

	err = repo.applyWalletDeltas(ctx, w.ID, -amount, amount)
	if err == nil {
		log.Info("funds reserved", "reserved", amount)
	}

	return err
}

func (repo SQLWalletRepository) RevertFundReservation(ctx context.Context, userId int, amount float64) error {
	log := util.FromContext(ctx)

	w, err := repo.loadWalletForUpdate(ctx, userId)
	if err != nil {
		return err
	}

	err = repo.applyWalletDeltas(ctx, w.ID, amount, -amount)
	if err == nil {
		log.Info("funds reverted", "amount", amount)
	}

	return err
}

func (repo SQLWalletRepository) loadWalletForUpdate(ctx context.Context, userId int) (*models.Wallet, error) {
	log := util.FromContext(ctx)

	row := repo.tx.QueryRowContext(ctx,
		`SELECT id, user_id, available_funds, reserved_funds
		 FROM brokerx.wallets
		 WHERE user_id = ?
		 FOR UPDATE`,
		userId,
	)

	var w models.Wallet
	if err := row.Scan(&w.ID, &w.UserId, &w.AvailableFunds, &w.ReservedFunds); err != nil {
		if err == sql.ErrNoRows {
			msg := "wallet does not exist"
			log.Error(msg, "error", err, "userId", userId)
			return nil, errors.New(msg)
		}
		log.Error("failed to fetch wallet", "error", err)
		return nil, err
	}

	return &w, nil
}

func (repo SQLWalletRepository) applyWalletDeltas(
	ctx context.Context,
	id string,
	availableDelta float64,
	reserveDelta float64,
) error {
	log := util.FromContext(ctx)

	_, err := repo.tx.ExecContext(ctx,
		`UPDATE brokerx.wallets
		 SET available_funds = available_funds + ?,
		     reserved_funds  = reserved_funds + ?
		 WHERE id = ?`,
		availableDelta, reserveDelta, id,
	)

	if err != nil {
		log.Error("failed to update wallet", "error", err)
	}

	return err
}



func (repo SQLWalletRepository) ReleaseFunds(ctx context.Context, deltas []models.WalletDelta) error {
	log := util.FromContext(ctx)

	if len(deltas) == 0 {
		return nil
	}

	// --- 1. Build the IN clause and lock all rows ---
	placeholders := make([]string, len(deltas))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	query := `
        SELECT id, user_id, available_funds, reserved_funds
        FROM brokerx.wallets
        WHERE user_id IN (` + strings.Join(placeholders, ",") + `)
        FOR UPDATE
    `

	args := make([]any, 0, len(deltas))
	for _, d := range deltas {
		args = append(args, d.Order.UserID)
	}

	rows, err := repo.tx.QueryContext(ctx, query, args...)
	if err != nil {
		log.Error("failed to lock rows", "error", err)
		return err
	}
	defer rows.Close()

	// --- 2. Load current values into map ---
	locked := make(map[int]models.Wallet)

	for rows.Next() {
		var lw models.Wallet

		if err := rows.Scan(&lw.ID, &lw.UserId, &lw.AvailableFunds, &lw.ReservedFunds); err != nil {
			return err
		}

		locked[lw.UserId] = lw
	}

	// --- 3. Validate constraints (simulate update) ---
	for _, d := range deltas {
		key := d.Order.UserID
		lw, ok := locked[key]
		if !ok {
			return fmt.Errorf("wallet not found for user=%d", d.Order.UserID)
		}

		if d.Order.Action == "buy" && 
			((d.Order.Type == "limit" && lw.ReservedFunds < d.Total) || 
			(d.Order.Type == "market" && lw.AvailableFunds < d.Total)) {
			return fmt.Errorf(
				"insufficient funds for user=%d : reserved=%f needed=%f",
				d.Order.UserID, lw.ReservedFunds, d.Total)
		}

		if d.Order.Action == "sell" {
			lw.AvailableFunds += d.Total
		} else if d.Order.Type == "limit" {
			lw.ReservedFunds -= d.Total
		} else if d.Order.Type == "market" {
			lw.AvailableFunds -= d.Total
		}

		locked[key] = lw
	}

	valueStrings := make([]string, 0, len(locked))
	valueArgs := make([]interface{}, 0, len(locked)*4)

	for _, lw := range locked {
		valueStrings = append(valueStrings, "(?, ?, ?, ?)")
		valueArgs = append(valueArgs, lw.ID, lw.UserId, lw.AvailableFunds, lw.ReservedFunds)
	}

	query = fmt.Sprintf(`
		INSERT INTO wallets (id, user_id, available_funds, reserved_funds)
		VALUES %s
		ON DUPLICATE KEY UPDATE
			reserved_funds = VALUES(reserved_funds),
			available_funds = VALUES(available_funds);
	`, strings.Join(valueStrings, ","))

	_, err = repo.tx.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		log.Error("Error executing wallet batch release", "error", err)
		return err
	}

	return nil
}

func (repo SQLWalletRepository) AddFunds(ctx context.Context, userId int, amount float64) error {

	log := util.FromContext(ctx)

	w, err := repo.loadWalletForUpdate(ctx, userId)
	if err != nil {
		return err
	}

	err = repo.applyWalletDeltas(ctx, w.ID, amount, 0)
	if err == nil {
		log.Info("funds reverted", "amount", amount)
	}

	return err
}

func (repo SQLWalletRepository) FindByUserId(ctx context.Context, userId int) (*models.Wallet, error) {
	log := util.FromContext(ctx)

	row := repo.DB.QueryRow("SELECT available_funds, reserved_funds FROM brokerx.wallets WHERE user_id=?", userId)

	var wallet models.Wallet
	e := row.Scan(&wallet.AvailableFunds, &wallet.ReservedFunds)
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
