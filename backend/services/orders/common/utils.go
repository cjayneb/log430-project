package common

import "errors"

type ctxKey string

const (
	CtxKeyJWT       ctxKey = "jwt"
	HeaderKeyUserId string = "X-User-ID"
	HeaderKeyAuth   string = "Authorization"
)

var (
	ErrBusinessRuleViolation = errors.New("business rule violation")
	ErrDependencyFailure     = errors.New("dependency failure")
)
