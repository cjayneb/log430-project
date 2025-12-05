package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type ctxKey string

const (
	CtxKeyJWT              ctxKey = "jwt"
	CtxKeyUserId           ctxKey = "user_id"
	CtxKeyTraceId          ctxKey = "traceId"
	HeaderKeyUserId        string = "X-User-ID"
	HeaderTraceId          string = "X-Trace-Id"
	HeaderKeyAuth          string = "Authorization"
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

var sharedTransport = &http.Transport{
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 100,
	IdleConnTimeout:     90 * time.Second,
	DisableCompression:  false,
}
var SharedHttpClient = &http.Client{
	Timeout:   5 * time.Second,
	Transport: sharedTransport,
}

func MakeAuthenticatedRequest(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	log := FromContext(ctx)
	ctx = context.WithoutCancel(ctx)
	jwt, ok := ctx.Value(CtxKeyJWT).(string)
	if !ok {
		return nil, fmt.Errorf("missing JWT in context")
	}
	userId, ok := ctx.Value(CtxKeyUserId).(string)
	if !ok {
		return nil, fmt.Errorf("missing UserId in context")
	}
	traceId, ok := ctx.Value(CtxKeyTraceId).(string)
	if !ok {
		return nil, fmt.Errorf("missing TraceId in context")
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request : %v", err)
	}
	req.Header.Set(HeaderKeyAuth, jwt)
	req.Header.Set(HeaderKeyUserId, userId)
	req.Header.Set(HeaderTraceId, traceId)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := SharedHttpClient.Do(req)
	if err != nil {
		msg := "error making request to user service"
		log.Error(msg, "error", err)
		return resp, errors.New(msg)
	}

	if resp.StatusCode != http.StatusOK {
		msg := "user service returned non-200 status"
		log.Error(msg, "status", resp.Status)
		return resp, errors.New(msg)
	}

	return resp, nil
}
