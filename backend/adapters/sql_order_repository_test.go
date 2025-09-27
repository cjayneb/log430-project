package adapters

import (
	"brokerx/models"
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var userId string = uuid.New().String()

func setupTestDBOrder(t *testing.T) (*sql.DB, func()) {
	// Run docker-compose.test.yml before executing this test
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		dbUrl = "root:root@tcp(127.0.0.1:3307)/brokerx?parseTime=true"
	} 
	defer os.Clearenv()
	
	log.Printf("Using DATABASE_URL: %s", dbUrl)
	db, err := sql.Open("mysql", dbUrl)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	_, err = db.Exec("DELETE FROM orders")
    require.NoError(t, err)

	_, err = db.Exec("DELETE FROM users")
    require.NoError(t, err)

    _, err = db.Query(`INSERT INTO users (id, email, password) 
                      VALUES (?, 'email', 'hashedpw')`, userId)
	require.NoError(t, err)

    _, err = db.Query(`INSERT INTO orders (user_id, symbol, type, action, quantity, unit_price, timing, status) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`, 
		userId, "AAPL", "buy", "market", 10, 150.00, "day", "open")
    require.NoError(t, err)

	cleanup := func() {
		db.Close()
	}
	return db, cleanup
}

func TestSQLOrderRepositoryIntegration(t *testing.T) {
	db, cleanup := setupTestDBOrder(t)
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
