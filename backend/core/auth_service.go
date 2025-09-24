package core

import (
	"brokerx/models"
	"brokerx/ports"
	"database/sql"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Repo ports.UserRepository
	PasswordAllowedRetries int
	PasswordLockDurationMinutes int
}

func (authService *AuthService) Authenticate(email, password string) (*models.User, error) {
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

func (authService *AuthService) lockUser(user *models.User) {
	user.FailedAttempts++
	if user.FailedAttempts >= authService.PasswordAllowedRetries {
		user.LockedUntil = sql.NullTime{
			Time: time.Now().Add(time.Duration(authService.PasswordLockDurationMinutes) * time.Minute), 
			Valid: true,
		}
	}

	err := authService.Repo.Update(user)
	if err != nil {
		log.Errorf("Failed to update user lock status: %v", err)
	}
}

func (authService *AuthService) resetLockout(user *models.User) {
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
