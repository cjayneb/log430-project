package ports

import "brokerx/user-service/models"

type UserRepository interface {
	FindByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Create(user *models.User) error
}
