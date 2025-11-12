package main

import (
	"brokerx/frontend-service/util"
	"context"
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/traceid"
	"github.com/google/uuid"
)

const TraceIDHeader = "X-Trace-Id"

type contextKey string

const TraceIdCtxKey contextKey = "traceId"

func main() {
	InitLogger("frontend-service")

	mux := http.NewServeMux()

	// Serve static files
	fs := http.FileServer(http.Dir("./public"))
	mux.Handle("/", fs)

	// Proxy API calls
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		var target *url.URL
		switch r.URL.Path {
		case "/api/user/auth/login", "/api/user/register":
			target, _ = url.Parse("http://user-service:8080")
		case "/api/order":
			target, _ = url.Parse("http://order-service:8080")
		case "/api/portfolio":
			target, _ = url.Parse("http://portfolio-service:8080")
		default:
			http.NotFound(w, r)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(target)

		traceID := r.Context().Value(TraceIdCtxKey)
		if traceIDStr, ok := traceID.(string); ok && traceIDStr != "" {
			r.Header.Set(TraceIDHeader, traceIDStr)
		}

		proxy.ServeHTTP(w, r)
	})

	handler := TraceMiddleware(mux)
	handler = httplog.RequestLogger(logger())(handler)
	handler = middleware.Recoverer(handler)


	slog.Info("Frontend service running on :8081")
	log.Fatal(http.ListenAndServe(":8081", handler))
}

func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		traceID := r.Header.Get(TraceIDHeader)
		if traceID == "" {
			traceID = uuid.New().String()
		}

		ctx = traceid.NewContext(ctx)
		ctx = context.WithValue(ctx, TraceIdCtxKey, traceID)

		reqLogger := slog.Default().With("traceId", traceID)
		ctx = util.WithLogger(ctx, reqLogger)

		httplog.LogEntrySetField(ctx, "traceId", slog.StringValue(traceID))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func InitLogger(service string) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	})

	logger := slog.New(handler).With("service", service)

	slog.SetDefault(logger)
}

func logger() *httplog.Logger {
	return httplog.NewLogger("frontend-service", httplog.Options{
		LogLevel:         slog.LevelDebug,
		RequestHeaders:   false,
		ResponseHeaders:  false,
		JSON:             false,
		Concise:          true,
		MessageFieldName: "message",
		LevelFieldName:   "severity",
		TimeFieldFormat:  time.RFC3339,
	})
}