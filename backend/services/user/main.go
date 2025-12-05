package main

import (
	dao_adapters "brokerx/user-service/adapters/dao"
	handler_adapters "brokerx/user-service/adapters/handlers"
	"brokerx/user-service/core"
	"brokerx/user-service/util"
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/traceid"
)

const TraceIDHeader = "X-Trace-Id"
type contextKey string
const TraceIdCtxKey contextKey = "traceId"

var config Config = Config{}

func main() {
	InitLogger("user-service")

	if err := config.LoadConfig(); err != nil {
		slog.Error("Config error", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting User Service", "port", config.Port)
	router := run()
	if err := http.ListenAndServe(":"+config.Port, router); err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
}

func run() http.Handler {
	userRepo := initDbConnection()

	authService := &core.AuthServiceImpl{
		Repo: userRepo,
		PasswordAllowedRetries: config.PasswordAllowedRetries,
		PasswordLockDurationMinutes: config.PasswordLockDurationMinutes,
	}
	userService := core.UserServiceImpl{Repo: userRepo}

	authHandler := &handler_adapters.AuthHandler{
		Service:   authService,
		JWTSecret: []byte(config.JWTSecret),
	}
	userHandler := handler_adapters.UserHandler{Service: &userService}

	r := chi.NewRouter()
	r.Use(httplog.RequestLogger(logger()))
	r.Use(middleware.Recoverer)
	r.Use(TraceMiddleware)

	r.Post("/api/user/register", authHandler.Register)
	r.Post("/api/user/auth/login", authHandler.Login)
	r.Get("/api/user/auth/verify", authHandler.VerifyToken)
	r.Get("/api/user/contact", userHandler.GetUserContactInfo)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		log := util.FromContext(r.Context())
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("{\"message\": \"User service OK\"}"))
		if err != nil {
			log.Error("Health check response error", "error", err)
		}
	})

	return r
}

func initDbConnection() *dao_adapters.SQLUserRepository {
	db, err := sql.Open("mysql", config.DBUrl)
	if err != nil {
		slog.Error("Db open error", "error", err)
		os.Exit(1)
	}

	db.SetMaxOpenConns(35)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Minute * 5)
	db.SetConnMaxIdleTime(time.Minute * 1)

	if err := db.Ping(); err != nil {
		slog.Error("Db ping error", "error", err)
	}
	return &dao_adapters.SQLUserRepository{DB: db}
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
	return httplog.NewLogger("user-service", httplog.Options{
		LogLevel:         slog.LevelInfo,
		RequestHeaders:   false,
		ResponseHeaders:  false,
		JSON:             false,
		Concise:          true,
		MessageFieldName: "message",
		LevelFieldName:   "severity",
		TimeFieldFormat:  time.RFC3339,
	})
}
