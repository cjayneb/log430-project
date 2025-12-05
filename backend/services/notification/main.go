package main

import (
	client_adapters "brokerx/notification-service/adapters/clients"
	dao_adapters "brokerx/notification-service/adapters/dao"
	handler_adapters "brokerx/notification-service/adapters/handlers"
	"brokerx/notification-service/core"
	"brokerx/notification-service/util"
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
)

var config Config = Config{}

func main() {
	InitLogger("notification-service")

	if err := config.LoadConfig(); err != nil {
		slog.Error("Config error", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting Notification Service", "port", config.Port)
	router := run()
	if err := http.ListenAndServe(":"+config.Port, router); err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
}

func run() http.Handler {
	notificationRepo := initDbConnection()
	userRepo := initRedisConnection()

	userServiceClient := client_adapters.NewUserServiceImpl(config.ApiGatewayBaseUrl)
	eventProducer := client_adapters.NewKafkaEventProducer(config.KafkaHost)

	notificationService := core.NotificationServiceImpl{
		NotificationRepo:  notificationRepo,
		UserRepository:    userRepo,
		UserServiceClient: userServiceClient,
		Producer:          eventProducer,
	}

	orderCompletedHandler := handler_adapters.OrderCompletedHandler{Service: &notificationService, Producer: eventProducer}
	eventConsumer := handler_adapters.NewKafkaEventConsumer(config.KafkaHost, config.KafkaGroupId, orderCompletedHandler)
	go eventConsumer.Start("NotificationEvents")

	r := chi.NewRouter()
	r.Use(httplog.RequestLogger(logger()))
	r.Use(middleware.Recoverer)

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

func initDbConnection() *dao_adapters.SQLNotificationRepository {
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
	return &dao_adapters.SQLNotificationRepository{DB: db}
}

func initRedisConnection() *dao_adapters.RedisUserRepository {
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: "",
		DB:       0,
	})

	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		slog.Error("Redis error", "error", err)
	}

	return &dao_adapters.RedisUserRepository{Rdb: client}
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
