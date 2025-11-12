package util

import (
	"context"
	"log/slog"
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
