package handler_adapters

import (
	"brokerx/user-service/core"
	"brokerx/user-service/models"
	"brokerx/user-service/util"
	"encoding/json"
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

type UserCreatedResponse struct {
	Message string `json:"message"`
	Email string `json:"email"`
	Status string `json:"status"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
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
	log := util.FromContext(r.Context())

	var creds LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		msg := "invalid JSON input"
		log.Error(msg, "error", err)
		util.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
		return
	}
	if err := validate.Struct(creds); err != nil {
		msg := "missing or invalid fields"
		log.Error(msg, "error", err)
		util.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
		return
	}

	user, err := handler.Service.Authenticate(r.Context(), creds.Email, creds.Password)
	if err != nil {
		msg := "invalid credentials"
		log.Error(msg, "error", err)
		util.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: msg})
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
		msg := "failed to sign token"
		log.Error(msg, "error", err)
		util.WriteJSON(w, http.StatusInternalServerError, ErrorResponse{ErrorMessage: msg})
		return
	}

	util.WriteJSON(w, http.StatusOK, LoginResponse{Token: tokenString})
}

func (handler *AuthHandler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	log := util.FromContext(r.Context())

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		msg := "missing Authorization header"
		log.Error(msg)
		util.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: msg})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if strings.Contains(tokenString, "Bearer ") {
		msg := "invalid Authorization header format"
		log.Error(msg)
		util.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{ErrorMessage: msg})
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
		msg := "invalid or expired token"
		log.Error(msg, "error", err)
		util.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
		return
	}

	w.Header().Set("X-User-Id", claims.UserID)
	util.WriteJSON(w, http.StatusOK, VerifyResponse{
		Valid:   true,
		UserId:  claims.UserID,
		Email:   claims.Email,
		Expires: *claims.ExpiresAt,
	})
}

func (handler *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	log := util.FromContext(r.Context())

	var newUser models.User
    if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		msg := "invalid JSON input"
		log.Error(msg, "error", err)
		util.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
        return
    }
    if err := validate.Struct(newUser); err != nil {
		msg := "missing or invalid fields"
		log.Error(msg, "error", err)
		util.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
        return
    }

    if err := handler.Service.Register(r.Context(), &newUser); err != nil {
        if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "registered") {
			msg := "email already registered"
			log.Error(msg)
			util.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
            return
        }
		msg := "error when registering user"
		log.Error(msg, "error", err)
		util.WriteJSON(w, http.StatusBadRequest, ErrorResponse{ErrorMessage: msg})
        return
    }

    util.WriteJSON(w, http.StatusCreated, UserCreatedResponse{
		Message: "user registered successfully",
		Email: newUser.Email,
		Status: newUser.Status,
	})
}
