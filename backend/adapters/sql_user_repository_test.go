package adapters

import (
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

var email string = "buyer@email.com"

func setupTestDB(t *testing.T) (*sql.DB, func()) {
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

func insertUserTestData(t *testing.T, db *sql.DB) {
	_, err := db.Query(`INSERT INTO users (id, email, password) 
                      VALUES (UUID(), ?, 'hashedpw')`, email)
    require.NoError(t, err)
}

func TestSQLUserRepositoryIntegration(t *testing.T) {
	db, cleanup := setupTestDB(t)
	insertUserTestData(t, db)
	defer cleanup()

	repo := &SQLUserRepository{DB: db}

	// --- FindByEmail ---
	expectedFailedAttempts := 0
	expectedLockedUntil := sql.NullTime{
		Time:  time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
		Valid: false,
	}
	user, err := repo.FindByEmail(email)
	require.NoError(t, err)
	require.Equal(t, email, user.Email)
	require.Equal(t, expectedFailedAttempts, user.FailedAttempts)
	require.WithinDuration(t, expectedLockedUntil.Time, user.LockedUntil.Time, time.Second)

	// --- Update ---
	expectedLockedUntil = sql.NullTime{
		Time:  time.Now().UTC(),
		Valid: true,
	}
	user.FailedAttempts = 2
	user.LockedUntil = expectedLockedUntil
	err = repo.Update(user)
	require.NoError(t, err)
	result, err := repo.FindByEmail(email)
	require.NoError(t, err)
	require.Equal(t, 2, result.FailedAttempts)
	require.WithinDuration(t, expectedLockedUntil.Time, result.LockedUntil.Time, time.Second)

	// --- FindByEmail non-existing user ---
	_, err = repo.FindByEmail("fakeemail")
	require.Error(t, err)
}
