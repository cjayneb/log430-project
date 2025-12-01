package util

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type ctxKey string

const (
	CtxKeyJWT       ctxKey = "jwt"
	CtxKeyUserId	ctxKey = "user_id"
	CtxKeyTraceId	ctxKey = "traceId"
	HeaderKeyUserId string = "X-User-ID"
	HeaderTraceId 	string = "X-Trace-Id"
	HeaderKeyAuth   string = "Authorization"
	AuthHeaderBearerPrefix string = "Bearer "
)

type contextKey struct{}

var key = contextKey{}

// WithLogger attaches a slog.Logger to a context.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, key, logger)
}

// FromContext retrieves the slog.Logger from context or returns slog.Default().
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(key).(*slog.Logger); ok && logger != nil {
		return logger
	}
	return slog.Default()
}

func FromEvent(traceId string) *slog.Logger {
	return slog.Default().With("traceId", traceId)
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("error when encoding JSON response : %v", err), http.StatusInternalServerError)
	}
}
