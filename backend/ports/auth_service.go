package ports

import "brokerx/models"

type AuthService interface {
    Authenticate(email, password string) (*models.User, error)
}
