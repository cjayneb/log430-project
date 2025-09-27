package adapters

import (
	"brokerx/models"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var symbol string = "AAPL"
var userId string = uuid.New().String()

func insertOrderTestData(t *testing.T, db *sql.DB) {
	_, err := db.Query(`INSERT INTO users (id, email, password) 
                      VALUES (?, 'email', 'hashedpw')`, userId)
	require.NoError(t, err)

    _, err = db.Query(`INSERT INTO orders (user_id, symbol, type, action, quantity, unit_price, timing, status) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`, 
		userId, symbol, "buy", "market", 10, 150.00, "day", "open")
    require.NoError(t, err)
}

func TestSQLOrderRepositoryIntegration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	insertOrderTestData(t, db)
	defer cleanup()

	repo := &SQLOrderRepository{DB: db}

	// --- Sucessfully create an order ---
	order := &models.Order{
		UserID:    userId,
		Symbol:    "AAPL",
		Type:      "buy",
		Action:    "market",
		Quantity:  10,
		UnitPrice: 150.00,
		Timing:    "day",
		Status:    "open",
	}

	id, err := repo.CreateOrder(order)

	require.Nil(t, err)
	require.Greater(t, id, 0)

	// --- Fail create an order ---
	badOrder := &models.Order{
		UserID:    userId,
		Symbol:    "AAPL",
		Type:      "buys",
		Action:    "market",
		Quantity:  10,
		UnitPrice: 150.00,
		Timing:    "day",
		Status:    "open",
	}

	id, err = repo.CreateOrder(badOrder)

	require.NotNil(t, err)
	require.Equal(t, 0, id)
}
