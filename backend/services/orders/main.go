package main

import (
	client_adapters "brokerx/order-service/adapters/clients"
	dao_adapters "brokerx/order-service/adapters/dao"
	handler_adapters "brokerx/order-service/adapters/handlers"
	"brokerx/order-service/core"
	"brokerx/order-service/ports"
	"context"
	"database/sql"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi/v5"
)

var config Config = Config{}

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Config error : %s", err)
	}

	// Create DAOs
	orderRepo, tm := initDbConnection()
	orderBook, execQueue := initRedisConnection()

	// Create external services
	matchingEngine := &ports.MatchineEngineImpl{}
	portfolioService := client_adapters.NewPortfolioServiceClient(config.ApiGatewayBaseUrl)
	marketDataProvider := client_adapters.NewMarketDataProvider(config.ApiGatewayBaseUrl)

	// Create core services
	complianceService := core.NewComplianceService(portfolioService, marketDataProvider)
	orderService := &core.OrderServiceImpl{
		Repo:              orderRepo,
		OrderBook:         orderBook,
		MatchingEngine:    matchingEngine,
	}

	// Create request handler
	orderHandler := handler_adapters.NewOrderHandler(orderService, complianceService)

	// Start async processes
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	initAsyncProcesses(ctx, orderService, orderBook, execQueue, tm)

	// Init router
	r := chi.NewRouter()
	r.Get("/api/order/", orderHandler.GetOrders)
	r.Post("/api/order/place", orderHandler.PlaceOrder)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("{\"message\": \"Order service OK\"}"))
		if err != nil {
			log.Errorf("Health check response error: %v", err)
		}
	})

	// Start service
	log.Println("Starting Order Service on port " + config.Port)
	http.ListenAndServe(":"+config.Port, r)
}

func initDbConnection() (*dao_adapters.SQLOrderRepository, *dao_adapters.SQLTransactionManager) {
	db, err := sql.Open("mysql", config.DBUrl)
	if err != nil {
		log.Fatalf("Db open error : %v", err)
	}

	db.SetMaxOpenConns(35)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Minute * 5)
	db.SetConnMaxIdleTime(time.Minute * 1)

	if err := db.Ping(); err != nil {
		log.Warnf("Db error : %s ", err)
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
		log.Warnf("Redis error : %v", err)
	}

	// TODO: Initialize RedisOrderBook with the database data
	return &dao_adapters.RedisOrderBook{Rdb: client}, &dao_adapters.RedisExecutionQueue{Rdb: client}
}

func initAsyncProcesses(
	ctx context.Context,
	orderService *core.OrderServiceImpl,
	orderBook *dao_adapters.RedisOrderBook,
	execQueue *dao_adapters.RedisExecutionQueue,
	tm *dao_adapters.SQLTransactionManager,
) {
	orderService.StartMatchingWorkers(config.NumberOfGoRoutines)
	core.StartDirtyOrderSync(
		ctx,
		time.Duration(config.DirtyOrderSyncIntervalInSeconds)*time.Second,
		config.DirtyOrderSyncBatchSize,
		orderBook,
		tm,
	)
	core.PersistOrdersAndExecutions(
		ctx,
		time.Duration(config.OrdersExecutionsPersistIntervalInMs)*time.Millisecond,
		config.OrdersPersistBatchSize,
		config.ExecutionsPersistBatchSize,
		orderBook,
		execQueue,
		tm,
	)
}
