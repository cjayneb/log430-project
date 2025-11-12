package dao_adapters

import (
	"brokerx/user-service/models"
	"brokerx/user-service/ports"
	"context"
	"database/sql"
)

type SQLUserRepository struct {
	DB *sql.DB
}

func (repo *SQLUserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
        INSERT INTO brokerx.users (email, password, first_name, last_name, status)
        VALUES (?, ?, ?, ?, ?)
    `
    _, err := repo.DB.Exec(query, user.Email, user.Password, user.FirstName, user.LastName, user.Status)
    return err
}

func (repo *SQLUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	row := repo.DB.QueryRow("SELECT id, email, password, failed_attempts, locked_until FROM brokerx.users WHERE email=?", email)

	var user models.User
	e := row.Scan(&user.ID, &user.Email, &user.Password, &user.FailedAttempts, &user.LockedUntil)
	if e != nil {
		return nil, e
	}

	return &user, nil
}

func (repo *SQLUserRepository) Update(ctx context.Context, user *models.User) error {
	_, e := repo.DB.Exec("UPDATE brokerx.users SET failed_attempts=?, locked_until=? WHERE email=?", user.FailedAttempts, user.LockedUntil, user.Email)
	return e
}

var _ ports.UserRepository = (*SQLUserRepository)(nil) // Ensure interface is implemented at compile time
