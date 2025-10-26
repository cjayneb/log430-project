package main

import (
	"brokerx/order-service/controllers"
	"brokerx/order-service/core"
	"brokerx/order-service/ports"
	"brokerx/order-service/repositories"
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

	orderRepo := initDbConnection()
	orderBook := initRedisConnection()

	matchingEngine := &ports.MatchineEngineImpl{}
	complianceService := &core.ComplianceService{
		PortfolioService:   &ports.PortfolioServiceImpl{},
		MarketDataProvider: &ports.MarketDataProviderImpl{},
	}
	orderService := &core.OrderService{
		Repo:              orderRepo,
		ComplianceService: complianceService,
		OrderBook:         orderBook,
		MatchingEngine:    matchingEngine,
	}
	orderHandler := controllers.NewOrderHandler(orderService)

	r := chi.NewRouter()
	r.Get("/api/order/", orderHandler.GetOrders)
	r.Post("/api/order/place", orderHandler.PlaceOrder)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("{\"message\": \"User service OK\"}"))
		if err != nil {
			log.Errorf("Health check response error: %v", err)
		}
	})

	log.Println("Starting Order Service on port " + config.Port)
	http.ListenAndServe(":"+config.Port, r)
}

func initDbConnection() *repositories.SQLOrderRepository {
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
	return &repositories.SQLOrderRepository{DB: db}
}

func initRedisConnection() *repositories.RedisOrderBook {
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
	return &repositories.RedisOrderBook{Rdb: client}
}
