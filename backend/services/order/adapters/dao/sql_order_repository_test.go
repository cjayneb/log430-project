package dao_adapters

import (
	"brokerx/order-service/models"
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

var ctx = context.Background()
var symbol string = "AAPL"
var userId int = 1

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// Run docker-compose.test.yml before executing this test
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		dbUrl = "root:root@tcp(127.0.0.1:3307)/brokerx?parseTime=true"
	}
	defer os.Clearenv()

	db, err := sql.Open("mysql", dbUrl)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	_, err = db.Exec("DELETE FROM orders")
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM positions")
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM wallets")
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM users")
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
	}
	return db, cleanup
}

func insertOrderTestData(t *testing.T, db *sql.DB) {
	_, err := db.Query(`INSERT INTO users (id, email, first_name, last_name, password) 
                      VALUES (?, email, 'hello', 'test', 'hashedpw')`, userId)
	require.NoError(t, err)

	_, err = db.Query(`INSERT INTO orders (user_id, symbol, type, action, quantity, remaining_quantity, unit_price, timing, status) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userId, symbol, "market", "buy", 10, 10, 150.00, "day", "open")
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
		Type:      "market",
		Action:    "buy",
		Quantity:  10,
		UnitPrice: 150.00,
		Timing:    "day",
		Status:    "open",
	}

	id, err := repo.Create(ctx, order)
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

	id, err = repo.Create(ctx, badOrder)
	require.NotNil(t, err)
	require.Equal(t, 0, id)
}
