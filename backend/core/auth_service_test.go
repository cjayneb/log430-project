package core

import (
	"brokerx/models"
	"bytes"
	"database/sql"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) FindByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func makeHashedPassword(pw string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(hash)
}

func makeUser(email, password string, failedAttempts int, lockedUntil sql.NullTime) *models.User {
	return &models.User{
		Email:    email,
		Password: makeHashedPassword(password),
		FailedAttempts: failedAttempts,
		LockedUntil: lockedUntil,
	}
}

func TestAuthenticateSuccess(t *testing.T) {
	email := "email"
	password := "password"
	user := makeUser(email, password, 0, sql.NullTime{Valid: false})
	repo := new(MockUserRepo)
	repo.On("FindByEmail", email).Return(user, nil)
	repo.On("Update", mock.Anything).Return(nil)
	service := &AuthService{
		Repo: repo,
		PasswordAllowedRetries: 3,
		PasswordLockDurationMinutes: 15,
	}

	result, err := service.Authenticate(email, password)

	require.NoError(t, err)
	require.Equal(t, user, result)
}

func TestAuthenticateUserNotFound(t *testing.T) {
	email := "email"
	password := "password"
	repo := new(MockUserRepo)
	repo.On("FindByEmail", email).Return(nil, sql.ErrNoRows)
	service := &AuthService{
		Repo: repo,
		PasswordAllowedRetries: 3,
		PasswordLockDurationMinutes: 15,
	}

	result, err := service.Authenticate(email, password)

	require.Nil(t, result)
	require.Error(t, err)
	require.Equal(t, "user not found", err.Error())
}

func TestAuthenticateUserLockUserUpdateFailure(t *testing.T) {
	expectedLog := "Failed to update user lock status: sql: connection is already closed"
	email := "email"
	password := "password"
	user := makeUser(email, password, 0, sql.NullTime{Valid: false})
	repo := new(MockUserRepo)
	repo.On("FindByEmail", email).Return(user, nil)
	repo.On("Update", mock.Anything).Return(sql.ErrConnDone)
	service := &AuthService{
		Repo: repo,
		PasswordAllowedRetries: 1,
		PasswordLockDurationMinutes: 15,
	}
	var buf bytes.Buffer
	originalOutput := log.StandardLogger().Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(originalOutput)

	result, err := service.Authenticate(email, "wrongpassword")
	logOutput := buf.String()

	require.Contains(t, logOutput, expectedLog)
	require.Nil(t, result)
	require.Error(t, err)
}

func TestAuthenticateResetLockoutUpdateFailure(t *testing.T) {
	expectedLog := "Failed to update user lock status: sql: connection is already closed"
	email := "email"
	password := "password"
	user := makeUser(email, password, 3, sql.NullTime{Valid: false})
	repo := new(MockUserRepo)
	repo.On("FindByEmail", email).Return(user, nil)
	repo.On("Update", mock.Anything).Return(sql.ErrConnDone)
	service := &AuthService{
		Repo: repo,
		PasswordAllowedRetries: 5,
		PasswordLockDurationMinutes: 5,
	}
	var buf bytes.Buffer
	originalOutput := log.StandardLogger().Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(originalOutput)

	result, err := service.Authenticate(email, password)
	logOutput := buf.String()

	require.Contains(t, logOutput, expectedLog)
	require.Equal(t, 0, result.FailedAttempts)
	require.Nil(t, err)
}

func TestAuthenticateInvalidPasswordTriggersLockout(t *testing.T) {
	email := "email"
	password := "password"
	user := makeUser(email, password, 0, sql.NullTime{Valid: false})
	repo := new(MockUserRepo)
	repo.On("FindByEmail", email).Return(user, nil)
	repo.On("Update", mock.Anything).Return(nil)
	service := &AuthService{
		Repo: repo,
		PasswordAllowedRetries: 1,
		PasswordLockDurationMinutes: 5,
	}

	result, err := service.Authenticate(email, "wrongpassword")

	require.Nil(t, result)
	require.Error(t, err)
	require.Equal(t, "invalid credentials", err.Error())
	require.True(t, user.LockedUntil.Valid)
}

func TestAuthenticateResetLockout(t *testing.T) {
	email := "email"
	password := "password"
	user := makeUser(email, password, 3, sql.NullTime{Valid: false})
	repo := new(MockUserRepo)
	repo.On("FindByEmail", email).Return(user, nil)
	repo.On("Update", mock.Anything).Return(nil)
	service := &AuthService{
		Repo: repo,
		PasswordAllowedRetries: 5,
		PasswordLockDurationMinutes: 5,
	}

	result, err := service.Authenticate(email, password)

	require.Equal(t, 0, result.FailedAttempts)
	require.Nil(t, err)
}

func TestAuthenticateAccountLocked(t *testing.T) {
	email := "email"
	password := "password"
	user := makeUser(email, password, 3, 
		sql.NullTime{
			Time:  time.Now().Add(10 * time.Minute),
			Valid: true,
		},
	)
	repo := new(MockUserRepo)
	repo.On("FindByEmail", email).Return(user, nil)
	service := &AuthService{
		Repo: repo,
		PasswordAllowedRetries: 3,
		PasswordLockDurationMinutes: 15,
	}

	result, err := service.Authenticate(email, password)

	require.Nil(t, result)
	require.Error(t, err)
	require.Equal(t, "account is locked. Try again later", err.Error())
}
