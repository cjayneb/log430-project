package core

import (
	"brokerx/models"
	"bytes"
	"database/sql"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
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
		Email:          email,
		Password:       makeHashedPassword(password),
		FailedAttempts: failedAttempts,
		LockedUntil:    lockedUntil,
	}
}

// ---------------------------
// Test Suite
// ---------------------------

type AuthServiceTestSuite struct {
	suite.Suite
	repo    *MockUserRepo
	service *AuthService
	email   string
	pass    string
}

func (s *AuthServiceTestSuite) SetupTest() {
	s.repo = new(MockUserRepo)
	s.service = &AuthService{
		Repo:                       s.repo,
		PasswordAllowedRetries:     3,
		PasswordLockDurationMinutes: 15,
	}
	s.email = "email"
	s.pass = "password"
}

// ---------------------------
// Tests
// ---------------------------

func (s *AuthServiceTestSuite) TestAuthenticateSuccess() {
	user := makeUser(s.email, s.pass, 0, sql.NullTime{Valid: false})
	s.repo.On("FindByEmail", s.email).Return(user, nil)
	s.repo.On("Update", mock.Anything).Return(nil)

	result, err := s.service.Authenticate(s.email, s.pass)

	s.Require().NoError(err)
	s.Equal(user, result)
}

func (s *AuthServiceTestSuite) TestAuthenticateUserNotFound() {
	s.repo.On("FindByEmail", s.email).Return(nil, sql.ErrNoRows)

	result, err := s.service.Authenticate(s.email, s.pass)

	s.Nil(result)
	s.Error(err)
	s.Equal("user not found", err.Error())
}

func (s *AuthServiceTestSuite) TestAuthenticateInvalidPasswordTriggersLockout() {
	user := makeUser(s.email, s.pass, 0, sql.NullTime{Valid: false})
	s.repo.On("FindByEmail", s.email).Return(user, nil)
	s.repo.On("Update", mock.Anything).Return(nil)
	s.service.PasswordAllowedRetries = 1

	result, err := s.service.Authenticate(s.email, "wrongpassword")

	s.Nil(result)
	s.Error(err)
	s.Equal("invalid credentials", err.Error())
	s.True(user.LockedUntil.Valid)
}

func (s *AuthServiceTestSuite) TestAuthenticateAccountLocked() {
	user := makeUser(s.email, s.pass, 0, sql.NullTime{
		Time:  time.Now().Add(10 * time.Minute),
		Valid: true,
	})
	s.repo.On("FindByEmail", s.email).Return(user, nil)

	result, err := s.service.Authenticate(s.email, s.pass)

	s.Nil(result)
	s.Error(err)
	s.Equal("account is locked. Try again later", err.Error())
}

func (s *AuthServiceTestSuite) TestAuthenticateUserLockUserUpdateFailure() {
	expectedLog := "Failed to update user lock status: sql: connection is already closed"
	user := makeUser(s.email, s.pass, 0, sql.NullTime{Valid: false})
	s.repo.On("FindByEmail", s.email).Return(user, nil)
	s.repo.On("Update", mock.Anything).Return(sql.ErrConnDone)
	s.service.PasswordAllowedRetries = 1

	var buf bytes.Buffer
	originalOutput := log.StandardLogger().Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(originalOutput)

	result, err := s.service.Authenticate(s.email, "wrongpassword")
	logOutput := buf.String()

	s.Contains(logOutput, expectedLog)
	s.Nil(result)
	s.Error(err)
}

func (s *AuthServiceTestSuite) TestAuthenticateResetLockout() {
	user := makeUser(s.email, s.pass, 3, sql.NullTime{Valid: false})
	s.repo.On("FindByEmail", s.email).Return(user, nil)
	s.repo.On("Update", mock.Anything).Return(nil)
	s.service.PasswordAllowedRetries = 5
	s.service.PasswordLockDurationMinutes = 5

	result, err := s.service.Authenticate(s.email, s.pass)

	s.Equal(0, result.FailedAttempts)
	s.NoError(err)
}

func (s *AuthServiceTestSuite) TestAuthenticateResetLockoutUpdateFailure() {
	expectedLog := "Failed to update user lock status: sql: connection is already closed"
	user := makeUser(s.email, s.pass, 3, sql.NullTime{Valid: false})
	s.repo.On("FindByEmail", s.email).Return(user, nil)
	s.repo.On("Update", mock.Anything).Return(sql.ErrConnDone)
	s.service.PasswordAllowedRetries = 5
	s.service.PasswordLockDurationMinutes = 5

	var buf bytes.Buffer
	originalOutput := log.StandardLogger().Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(originalOutput)

	result, err := s.service.Authenticate(s.email, s.pass)
	logOutput := buf.String()

	s.Contains(logOutput, expectedLog)
	s.Equal(0, result.FailedAttempts)
	s.NoError(err)
}

// ---------------------------
// Run the suite
// ---------------------------
func TestAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}
