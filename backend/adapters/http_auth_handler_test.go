package adapters

import (
	"brokerx/models"
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockAuthService implements core.AuthService interface for tests
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Authenticate(email, password string) (*models.User, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// ---------------------------
// Test Suite
// ---------------------------

type HttpAuthHandlerTestSuite struct {
	suite.Suite
	service *MockAuthService
	handler *AuthHandler
}

func (s *HttpAuthHandlerTestSuite) SetupTest() {
	s.service = new(MockAuthService)
	s.handler = &AuthHandler{Service: s.service, IsProduction: false}
}

func (s *HttpAuthHandlerTestSuite) TestLoginSuccess() {
	user := &models.User{Email: "test@x.com", Password: "hashed", FailedAttempts: 0, LockedUntil: sql.NullTime{Valid: false}}
	s.service.On("Authenticate", "test@x.com", "pw").Return(user, nil)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("email=test@x.com&password=pw"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	s.handler.Login(w, req)

	res := w.Result()
	defer res.Body.Close()

	s.Equal(http.StatusFound, res.StatusCode)
	s.Equal("/", res.Header.Get("Location"))
}

func (s *HttpAuthHandlerTestSuite) TestLoginBadRequest() {
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("email=&password="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	s.handler.Login(w, req)

	s.Equal(http.StatusBadRequest, w.Result().StatusCode)
}

func (s *HttpAuthHandlerTestSuite) TestLoginUnauthorized() {
	s.service.On("Authenticate", "bad@x.com", "wrong").Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("email=bad@x.com&password=wrong"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	s.handler.Login(w, req)

	s.Equal(http.StatusUnauthorized, w.Result().StatusCode)
}

func (s *HttpAuthHandlerTestSuite) TestMiddlewareUnauthenticated() {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	// Wrap a dummy handler with middleware
	protected := s.handler.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	protected.ServeHTTP(w, req)

	s.Equal(http.StatusFound, w.Result().StatusCode)
	s.Equal("/login", w.Result().Header.Get("Location"))
}

// ---------------------------
// Run the suite
// ---------------------------
func TestHttpAuthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HttpAuthHandlerTestSuite))
}
