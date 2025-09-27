package adapters

import (
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var availableFunds float64 = 1000.0
var fundsOnHold float64 = 150.0

func insertWalletTestData(t *testing.T, db *sql.DB) {
	_, err := db.Query(`INSERT INTO users (id, email, password) 
                      VALUES (?, 'email', 'hashedpw')`, userId)
	require.NoError(t, err)

    _, err = db.Query(`INSERT INTO wallets (id, user_id, available_funds, funds_on_hold) VALUES(?, ?, ?, ?)`, 
		uuid.New().String(), userId, availableFunds, fundsOnHold)
    require.NoError(t, err)
}

func TestSQLWalletRepositoryIntegration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	insertWalletTestData(t, db)
	defer cleanup()

	repo := &SQLWalletRepository{DB: db}

	// --- FindByUserId ---
	wallet, err := repo.FindByUserId(userId)
	require.NoError(t, err)
	require.Equal(t, availableFunds, wallet.AvailableFunds)
	require.Equal(t, fundsOnHold, wallet.OnHoldFunds)

	// --- FindByUserId not found ---
	wallet, err = repo.FindByUserId("non existent user id")
	require.ErrorIs(t, err, sql.ErrNoRows)
	require.Nil(t, wallet)
}
