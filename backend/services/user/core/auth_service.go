package core

import (
	"brokerx/user-service/models"
	"brokerx/user-service/ports"
	"database/sql"
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Authenticate(email, password string) (*models.User, error)
	Register(user *models.User) error
}

type AuthServiceImpl struct {
	Repo                        ports.UserRepository
	PasswordAllowedRetries      int
	PasswordLockDurationMinutes int
}

func (authService *AuthServiceImpl) Register(user *models.User) error {
    existing, err := authService.Repo.FindByEmail(user.Email)
    if err == nil && existing != nil {
        return errors.New("email already registered")
    }

    hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        return fmt.Errorf("failed to hash password: %w", err)
    }
    user.Password = string(hashed)
    user.Status = "pending"
	// TODO: Send confirmation email via notification service

    return authService.Repo.Create(user)
}

func (authService *AuthServiceImpl) Authenticate(email, password string) (*models.User, error) {
	user, e := authService.Repo.FindByEmail(email)
	if e != nil {
		return nil, errors.New("user not found")
	}

	if user.LockedUntil.Valid && user.LockedUntil.Time.After(time.Now()) {
		return nil, errors.New("account is locked. Try again later")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		authService.lockUser(user)
		return nil, errors.New("invalid credentials")
	}

	authService.resetLockout(user)
	return user, nil
}

func (authService *AuthServiceImpl) lockUser(user *models.User) {
	user.FailedAttempts++
	if user.FailedAttempts >= authService.PasswordAllowedRetries {
		user.LockedUntil = sql.NullTime{
			Time:  time.Now().Add(time.Duration(authService.PasswordLockDurationMinutes) * time.Minute),
			Valid: true,
		}
	}

	err := authService.Repo.Update(user)
	if err != nil {
		log.Errorf("Failed to update user lock status: %v", err)
	}
}

func (authService *AuthServiceImpl) resetLockout(user *models.User) {
	if user.FailedAttempts == 0 {
		return
	}
	user.FailedAttempts = 0
	user.LockedUntil = sql.NullTime{Valid: false}

	err := authService.Repo.Update(user)
	if err != nil {
		log.Errorf("Failed to update user lock status: %v", err)
	}
}

var _ AuthService = (*AuthServiceImpl)(nil) // Ensure interface is implemented at compile time
