package adapters

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var quantity int = 1000
var unitPrice float64 = 150.0

func insertPositionTestData(t *testing.T, db *sql.DB) {
	_, err := db.Query(`INSERT INTO users (id, email, password) 
                      VALUES (?, 'email', 'hashedpw')`, userId)
	require.NoError(t, err)

    _, err = db.Query(`INSERT INTO positions (user_id, symbol, quantity, unit_price) VALUES(?, ?, ?, ?)`, 
		userId, symbol, quantity, unitPrice)
    require.NoError(t, err)
}

func TestSQLPositionRepositoryIntegration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	insertPositionTestData(t, db)
	defer cleanup()

	repo := &SQLPositionRepository{DB: db}

	// --- FindByUserIdAndSymbol ---
	positions, err := repo.FindByUserIdAndSymbol(userId, symbol)
	require.NoError(t, err)
	require.Equal(t, 1, len(positions))
	require.Equal(t, symbol, positions[0].Symbol)
	require.Equal(t, quantity, positions[0].Quantity)
	require.Equal(t, unitPrice, positions[0].UnitPrice)

	// --- FindByUserIdAndSymbol No positions ---
	positions, err = repo.FindByUserIdAndSymbol(userId, "stockThatUserDoesntOwn")
	require.NoError(t, err)
	require.Equal(t, 0, len(positions))

	// --- FindByUserIdAndSymbol connection error ---
	db, mock, _ := sqlmock.New()
	repo = &SQLPositionRepository{DB: db}
	mock.ExpectQuery(".*").
		WillReturnError(sql.ErrConnDone)

	positions, err = repo.FindByUserIdAndSymbol(userId, symbol)
	require.Nil(t, positions)
	require.ErrorIs(t, err, sql.ErrConnDone)

	// --- FindByUserIdAndSymbol scan error ---
	rows := sqlmock.NewRows([]string{"symbol", "quantity", "unit_price"}).
		AddRow("AAPL", 10, "bad-data")
	mock.ExpectQuery(".*").WillReturnRows(rows)

	positions, err = repo.FindByUserIdAndSymbol(userId, symbol)
	assert.Nil(t, positions)
	assert.Error(t, err)
}
