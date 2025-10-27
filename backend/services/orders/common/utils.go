package common

import "errors"

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
