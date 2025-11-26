package main

import (
	client_adapters "brokerx/order-service/adapters/clients"
	dao_adapters "brokerx/order-service/adapters/dao"
	handler_adapters "brokerx/order-service/adapters/handlers"
	"brokerx/order-service/common"
	"brokerx/order-service/core"
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/traceid"
)

var config Config = Config{}

func main() {
	InitLogger("order-service")

	if err := config.LoadConfig(); err != nil {
		slog.Error("Config error", "error", err)
		os.Exit(1)
	}

	// Create DAOs
	orderRepo, tm := initDbConnection()
	orderBook, execQueue := initRedisConnection()

	// Create external services
	matchingEngine := client_adapters.NewMatchineEngine(config.MatchingServiceBaseUrl)
	eventProducer := client_adapters.NewKafkaEventProducer("kafka:9092")
	portfolioService := client_adapters.NewPortfolioServiceClient(config.PortfolioServiceBaseUrl)
	marketDataProvider := client_adapters.NewMarketDataProvider(config.MarketDataServiceBaseUrl)

	// Create core services
	complianceService := core.NewComplianceService(portfolioService, marketDataProvider)
	orderService := &core.OrderServiceImpl{
		Repo:              orderRepo,
		OrderBook:         orderBook,
		MatchingEngine:    matchingEngine,
		EventProducer: eventProducer,
	}

	// Create request/event handlers
	orderHandler := handler_adapters.NewOrderHandler(orderService, complianceService)
	orderCreatedHandler := handler_adapters.OrderCreatedHandler{ComplianceService: complianceService, Producer: eventProducer}
	orderFailedHandler := handler_adapters.OrderComplianceFailedHandler{OrderService: orderService, Producer: eventProducer}
	eventConsumer := handler_adapters.NewKafkaEventConsumer("kafka:9092", "group1", orderCreatedHandler, orderFailedHandler)

	// Start async processes
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	initAsyncProcesses(ctx, orderBook, execQueue, eventConsumer, tm)

	// Init router
	r := chi.NewRouter()
	r.Use(httplog.RequestLogger(logger()))
	r.Use(middleware.Recoverer)
	r.Use(TraceMiddleware)

	r.Get("/api/order/", orderHandler.GetOrders)
	r.Post("/api/order/place", orderHandler.PlaceOrder)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		log := common.FromContext(r.Context())
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("{\"message\": \"Order service OK\"}"))
		if err != nil {
			log.Error("Health check response error", "error", err)
		}
	})

	slog.Info("Starting Order Service", "port", config.Port)
	if err := http.ListenAndServe(":"+config.Port, r); err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
}

func initDbConnection() (*dao_adapters.SQLOrderRepository, *dao_adapters.SQLTransactionManager) {
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
		slog.Warn("Db ping error", "error", err)
	}
	return &dao_adapters.SQLOrderRepository{DB: db}, &dao_adapters.SQLTransactionManager{DB: db}
}

func initRedisConnection() (*dao_adapters.RedisOrderBook, *dao_adapters.RedisExecutionQueue) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: "",
		DB:       0,
	})

	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		slog.Warn("Redis ping error", "error", err)
	}

	// TODO: Initialize RedisOrderBook with the database data
	return &dao_adapters.RedisOrderBook{Rdb: client}, &dao_adapters.RedisExecutionQueue{Rdb: client}
}

func initAsyncProcesses(
	ctx context.Context,
	orderBook *dao_adapters.RedisOrderBook,
	execQueue *dao_adapters.RedisExecutionQueue,
	eventConsumer *handler_adapters.KafkaEventConsumer,
	tm *dao_adapters.SQLTransactionManager,
) {
	// core.StartDirtyOrderSync(
	// 	ctx,
	// 	time.Duration(config.DirtyOrderSyncIntervalInSeconds)*time.Second,
	// 	config.DirtyOrderSyncBatchSize,
	// 	orderBook,
	// 	tm,
	// )
	// core.PersistOrdersAndExecutions(
	// 	ctx,
	// 	time.Duration(config.OrdersExecutionsPersistIntervalInMs)*time.Millisecond,
	// 	config.OrdersPersistBatchSize,
	// 	config.ExecutionsPersistBatchSize,
	// 	orderBook,
	// 	execQueue,
	// 	tm,
	// )
	go eventConsumer.Start("OrderEvents")
}

func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		traceID := r.Header.Get(common.HeaderTraceId)
		if traceID == "" {
			traceID = uuid.New().String()
		}

		ctx = traceid.NewContext(ctx)
		ctx = context.WithValue(ctx, common.CtxKeyTraceId, traceID)

		reqLogger := slog.Default().With("traceId", traceID)
		ctx = common.WithLogger(ctx, reqLogger)

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
	return httplog.NewLogger("order-service", httplog.Options{
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
