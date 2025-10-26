package ports

import "brokerx/user-service/models"

type AuthService interface {
	Authenticate(email, password string) (*models.User, error)
}
