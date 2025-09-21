package core

import (
	"brokerx/models"
	"brokerx/ports"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Repo ports.UserRepository
}

func (authService *AuthService) Authenticate(email, password string) (*models.User, error) {
	user, e := authService.Repo.FindByEmail(email)
	if e != nil {
		return nil, errors.New("user not found")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}