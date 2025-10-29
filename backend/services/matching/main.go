package main

import (
	dao_adapters "brokerx/matching-service/adapters/dao"
	handler_adapters "brokerx/matching-service/adapters/handlers"
	"brokerx/matching-service/core"
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

var config Config = Config{}

func main() {
	log.Println("Starting Matching Service on port " + config.Port)
	router := run()
	if err := http.ListenAndServe(":"+config.Port, router); err != nil {
		log.Fatalf("Server error : %s", err)
	}
}

func run() http.Handler {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Config error : %s", err)
	}

	orderBook, execQueue := initRedisConnection()

	matchingEngine := &core.MatchingEngineImpl{
		OrderBook:      orderBook,
		ExecutionQueue: execQueue,
	}
	matchingHandler := &handler_adapters.MatchingHandler{
		MatchingEngine: matchingEngine,
	}

	matchingEngine.StartMatchingWorkers(config.NumberOfGoRoutines)

	r := chi.NewRouter()
	r.Post("/api/matching/", matchingHandler.SubmitOrder)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("{\"message\": \"Matching service OK\"}"))
		if err != nil {
			log.Errorf("Health check response error: %v", err)
		}
	})
	return r
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
