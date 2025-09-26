package adapters

import (
	"brokerx/models"
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var LOGIN_ENDPOINT string = "/auth/login"

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

type FailingStore struct{}
func (f *FailingStore) Get(r *http.Request, name string) (*sessions.Session, error) {
    return sessions.NewSession(f, name), nil
}

func (f *FailingStore) New(r *http.Request, name string) (*sessions.Session, error) {
    return sessions.NewSession(f, name), nil
}
func (f *FailingStore) Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error {
    return errors.New("failed to save session")
}

// ---------------------------
// Test Suite
// ---------------------------

type HttpAuthHandlerTestSuite struct {
	suite.Suite
	mockService *MockAuthService
	handler *AuthHandler
}

func (s *HttpAuthHandlerTestSuite) SetupTest() {
	s.mockService = new(MockAuthService)
	s.handler = &AuthHandler{Service: s.mockService, SessionStore: sessions.NewCookieStore([]byte("very-secret-key")), IsProduction: false}
}

func (s *HttpAuthHandlerTestSuite) TestLoginSuccess() {
	user := &models.User{Email: "test@x.com", Password: "hashed", FailedAttempts: 0, LockedUntil: sql.NullTime{Valid: false}}
	s.mockService.On("Authenticate", "test@x.com", "pw").Return(user, nil)

	req := httptest.NewRequest(http.MethodPost, LOGIN_ENDPOINT, bytes.NewBufferString("email=test@x.com&password=pw"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	s.handler.Login(w, req)
	res := w.Result()
	defer res.Body.Close()

	s.Equal(http.StatusFound, res.StatusCode)
	s.Equal("/", res.Header.Get("Location"))
	s.Equal("brokerx-session", res.Cookies()[0].Name)
}

func (s *HttpAuthHandlerTestSuite) TestLoginBadRequest() {
	req := httptest.NewRequest(http.MethodPost, LOGIN_ENDPOINT, bytes.NewBufferString("email=&password="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	s.handler.Login(w, req)

	s.Equal(http.StatusBadRequest, w.Result().StatusCode)
}

func (s *HttpAuthHandlerTestSuite) TestLoginUnauthorized() {
	s.mockService.On("Authenticate", "bad@x.com", "wrong").Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodPost, LOGIN_ENDPOINT, bytes.NewBufferString("email=bad@x.com&password=wrong"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	s.handler.Login(w, req)

	s.Equal(http.StatusUnauthorized, w.Result().StatusCode)
}

func (s *HttpAuthHandlerTestSuite) TestMiddlewareUnauthenticated() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	protected := s.handler.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	protected.ServeHTTP(w, req)

	s.Equal(http.StatusFound, w.Result().StatusCode)
	s.Equal("/login", w.Result().Header.Get("Location"))
}

func (s *HttpAuthHandlerTestSuite) TestMiddlewareAuthenticated() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	session, _ := s.handler.SessionStore.Get(req, "brokerx-session")
    session.Values["user_id"] = "test@example.com"
    require.NoError(s.T(), session.Save(req, w))

	req.AddCookie(w.Result().Cookies()[0])

	protected := s.handler.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	protected.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Result().StatusCode)
}

func (s *HttpAuthHandlerTestSuite) TestInitSessionFailure() {
	user := &models.User{Email: "test@x.com", Password: "hashed", FailedAttempts: 0, LockedUntil: sql.NullTime{Valid: false}}
	s.mockService.On("Authenticate", "test@x.com", "pw").Return(user, nil)

    s.handler.SessionStore = &FailingStore{}
    req := httptest.NewRequest(http.MethodPost, LOGIN_ENDPOINT, bytes.NewBufferString("email=test@x.com&password=pw"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    w := httptest.NewRecorder()

    s.handler.Login(w, req)
	res := w.Result()
	defer res.Body.Close()

	s.Equal(http.StatusInternalServerError, res.StatusCode)
	s.Contains(w.Body.String(), "failed to save session")
}


// ---------------------------
// Run the suite
// ---------------------------
func TestHttpAuthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HttpAuthHandlerTestSuite))
}
