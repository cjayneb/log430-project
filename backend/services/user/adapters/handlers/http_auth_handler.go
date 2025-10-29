package handler_adapters

import (
	"brokerx/user-service/core"
	"brokerx/user-service/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const USER_ID_KEY contextKey = "user_id"

type AuthHandler struct {
	Service   core.AuthService
	JWTSecret []byte
}

type LoginResponse struct {
	Token string `json:"token"`
}

type VerifyResponse struct {
	Valid   bool            `json:"valid"`
	UserId  string          `json:"user_id"`
	Email   string          `json:"email"`
	Expires jwt.NumericDate `json:"expires"`
}

type ErrorResponse struct {
	ErrorMessage string `json:"errorMessage"`
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

var validate = validator.New()

func (handler *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds models.User
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: "invalid JSON input"})
		return
	}
	if err := validate.Struct(creds); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: fmt.Sprintf("missing or invalid fields: %v", err)})
		return
	}

	user, err := handler.Service.Authenticate(creds.Email, creds.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: "invalid credentials"})
		return
	}

	expiration := time.Now().Add(2 * time.Hour)
	claims := &Claims{
		UserID: strconv.Itoa(user.ID),
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "brokerx-user-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(handler.JWTSecret)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{ErrorMessage: "failed to sign token"})
		return
	}

	writeJSON(w, http.StatusOK, LoginResponse{Token: tokenString})
}

func (handler *AuthHandler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: "missing Authorization header"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if strings.Contains(tokenString, "Bearer ") {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: "invalid Authorization header format"})
		return
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return handler.JWTSecret, nil
	})
	if err != nil || !token.Valid {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: "invalid or expired token"})
		return
	}

	w.Header().Set("X-User-Id", claims.UserID)
	writeJSON(w, http.StatusOK, VerifyResponse{
		Valid:   true,
		UserId:  claims.UserID,
		Email:   claims.Email,
		Expires: *claims.ExpiresAt,
	})
}

func (handler *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(501)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("error when encoding JSON response : %v", err), http.StatusInternalServerError)
	}
}
