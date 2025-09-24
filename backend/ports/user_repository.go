package ports

import "brokerx/models"

type UserRepository interface {
	FindByEmail(email string) (*models.User, error)
	Update(user *models.User) error
}
