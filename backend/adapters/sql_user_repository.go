package adapters

import (
	"brokerx/models"
	"brokerx/ports"
	"database/sql"
)

type SQLUserRepository struct {
	DB *sql.DB
}

func (repo * SQLUserRepository) FindByEmail(email string) (*models.User, error) {
	row := repo.DB.QueryRow("SELECT id, email, password, failed_attempts, locked_until FROM brokerx.users WHERE email=?", email)

	var user models.User
	e := row.Scan(&user.ID, &user.Email, &user.Password, &user.FailedAttempts, &user.LockedUntil)
	if e != nil {
		return nil, e
	}

	return &user, nil
}

func (repo * SQLUserRepository) Update(user *models.User) error {
	_, e := repo.DB.Exec("UPDATE brokerx.users SET failed_attempts=?, locked_until=? WHERE email=?", user.FailedAttempts, user.LockedUntil, user.Email)
	return e
}

var _ ports.UserRepository = (*SQLUserRepository)(nil) // Ensure interface is implemented at compile time