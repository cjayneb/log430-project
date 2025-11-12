package core

import (
	"brokerx/user-service/models"
	"brokerx/user-service/ports"
	"brokerx/user-service/util"
	"context"
	"database/sql"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Authenticate(ctx context.Context, email, password string) (*models.User, error)
	Register(ctx context.Context, user *models.User) error
}

type AuthServiceImpl struct {
	Repo                        ports.UserRepository
	PasswordAllowedRetries      int
	PasswordLockDurationMinutes int
}

func (authService *AuthServiceImpl) Register(ctx context.Context, user *models.User) error {
	log := util.FromContext(ctx)

    existing, err := authService.Repo.FindByEmail(ctx, user.Email)
    if err == nil && existing != nil {
		msg := "email already regostered"
		log.Error(msg)
        return errors.New(msg)
    }

    hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
		msg := "failed to hash password"
		log.Error(msg, "error", err)
        return errors.New(msg)
    }
    user.Password = string(hashed)
    user.Status = "pending"
	// TODO: Send confirmation email via notification service

    return authService.Repo.Create(ctx, user)
}

func (authService *AuthServiceImpl) Authenticate(ctx context.Context, email, password string) (*models.User, error) {
	user, e := authService.Repo.FindByEmail(ctx, email)
	if e != nil {
		msg := "user not found"
		log.Error(msg, "error", e)
        return nil, errors.New(msg)
	}

	if user.LockedUntil.Valid && user.LockedUntil.Time.After(time.Now()) {
		msg := "account is locked. Try again later"
		log.Error(msg)
        return nil, errors.New(msg)
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		authService.lockUser(ctx, user)
		msg := "invalid credentials"
		log.Error(msg)
        return nil, errors.New(msg)
	}

	authService.resetLockout(ctx, user)
	return user, nil
}

func (authService *AuthServiceImpl) lockUser(ctx context.Context, user *models.User) {
	log := util.FromContext(ctx)

	user.FailedAttempts++
	if user.FailedAttempts >= authService.PasswordAllowedRetries {
		user.LockedUntil = sql.NullTime{
			Time:  time.Now().Add(time.Duration(authService.PasswordLockDurationMinutes) * time.Minute),
			Valid: true,
		}
	}

	err := authService.Repo.Update(ctx, user)
	if err != nil {
		log.Error("Failed to update user lock status", "error", err)
	}
}

func (authService *AuthServiceImpl) resetLockout(ctx context.Context, user *models.User) {
	log := util.FromContext(ctx)

	if user.FailedAttempts == 0 {
		return
	}
	user.FailedAttempts = 0
	user.LockedUntil = sql.NullTime{Valid: false}

	err := authService.Repo.Update(ctx, user)
	if err != nil {
		log.Error("Failed to update user lock status", "error", err)
	}
}

var _ AuthService = (*AuthServiceImpl)(nil) // Ensure interface is implemented at compile time
