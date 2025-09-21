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
	row := repo.DB.QueryRow("SELECT id, email, password FROM brokerx.users WHERE email=?", email)

	var user models.User
	e := row.Scan(&user.ID, &user.Email, &user.Password)
	if e != nil {
		return nil, e
	}

	return &user, nil
}

var _ ports.UserRepository = (*SQLUserRepository)(nil) // Ensure interface is implemented at compile time