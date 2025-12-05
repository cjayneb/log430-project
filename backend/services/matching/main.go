package main

import (
	client_adapters "brokerx/matching-service/adapters/clients"
	dao_adapters "brokerx/matching-service/adapters/dao"
	handler_adapters "brokerx/matching-service/adapters/handlers"
	"brokerx/matching-service/core"
	"brokerx/matching-service/util"
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
	"github.com/redis/go-redis/v9"
)

const TraceIDHeader = "X-Trace-Id"
type contextKey string
const TraceIdCtxKey contextKey = "traceId"

var config Config = Config{}

func main() {
	InitLogger("matching-service")

	if err := config.LoadConfig(); err != nil {
		slog.Error("Config error", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting Matching Service", "port", config.Port)
	router := run()
	if err := http.ListenAndServe(":"+config.Port, router); err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
}

func run() http.Handler {
	orderBook:= initRedisConnection()

	eventProducer := client_adapters.NewKafkaEventProducer("kafka:9092")
	matchingEngine := &core.MatchingEngineImpl{
		OrderBook:      orderBook,
		Producer: eventProducer,
	}
	orderValidatedHandler := handler_adapters.OrderValidatedHandler{MatchingService: matchingEngine, Producer: eventProducer}
	orderOpenHandler := handler_adapters.OrderOpenHandler{OrderBook: orderBook}
	eventConsumer := handler_adapters.NewKafkaEventConsumer(config.KafkaHost, config.KafkaGroupId, orderValidatedHandler, orderOpenHandler)

	//matchingEngine.StartMatchingWorkers(config.NumberOfGoRoutines)
	eventConsumer.Start("MatchingEvents")

	r := chi.NewRouter()
	r.Use(httplog.RequestLogger(logger()))
	r.Use(middleware.Recoverer)
	r.Use(TraceMiddleware)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("{\"message\": \"Matching service OK\"}"))
		if err != nil {
			slog.Error("Health check response error", "error", err)
		}
	})
	return r
}

func initRedisConnection() *dao_adapters.RedisOrderBook {
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

	return &dao_adapters.RedisOrderBook{Rdb: client}
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
	return httplog.NewLogger("matching-service", httplog.Options{
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
