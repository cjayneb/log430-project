package main

import (
	handler_adapters "brokerx/market-data-service/adapters/handlers"
	"brokerx/market-data-service/core"
	"brokerx/market-data-service/util"
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/traceid"
	"github.com/google/uuid"
)

const TraceIDHeader = "X-Trace-Id"
type contextKey string
const TraceIdCtxKey contextKey = "traceId"

var config Config = Config{}

func main() {
	InitLogger("market-data-service")

	if err := config.LoadConfig(); err != nil {
		slog.Error("Config error", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting Market data Service", "port", config.Port)
	router := run()
	if err := http.ListenAndServe(":"+config.Port, router); err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
}

func run() http.Handler {
	marketDataService := core.NewMarketDataServiceImpl(config.ResourcePath)
	marketDataHandler := handler_adapters.MarketDataHandler{Service: marketDataService}

	r := chi.NewRouter()
	r.Use(httplog.RequestLogger(logger()))
	r.Use(middleware.Recoverer)
	r.Use(TraceMiddleware)

	r.Get("/api/market/stock/price", marketDataHandler.GetStockPrice)
	r.Get("/api/market/stock/instrument", marketDataHandler.GetInstrument)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("{\"message\": \"Market data service OK\"}"))
		if err != nil {
			slog.Error("Health check response error", "error", err)
		}
	})
	return r
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
	return httplog.NewLogger("market-data-service", httplog.Options{
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
