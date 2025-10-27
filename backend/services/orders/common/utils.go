package common

import (
	"errors"
	"net/http"
	"time"
)

type ctxKey string

const (
	CtxKeyJWT       ctxKey = "jwt"
	CtxKeyUserId	ctxKey = "user_id"
	HeaderKeyUserId string = "X-User-ID"
	HeaderKeyAuth   string = "Authorization"
	AuthHeaderBearerPrefix string = "Bearer "
)

var (
	ErrBusinessRuleViolation = errors.New("business rule violation")
	ErrDependencyFailure     = errors.New("dependency failure")
)

var sharedTransport = &http.Transport{
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 100,
	IdleConnTimeout:     90 * time.Second,
}
var SharedHttpClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: sharedTransport,
}
