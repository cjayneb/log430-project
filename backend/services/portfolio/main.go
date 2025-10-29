package main

import (
	dao_adapters "brokerx/portfolio-service/adapters/dao"
	handler_adapters "brokerx/portfolio-service/adapters/handlers"
	"brokerx/portfolio-service/core"
	"database/sql"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi/v5"
)

var config Config = Config{}

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Config error : %s", err)
	}

	walletRepo, positionsRepo := initDbConnection()

	portfolioService := &core.PortfolioServiceImpl{
		WalletRepo:    walletRepo,
		PositionsRepo: positionsRepo,
	}
	portfolioHandler := handler_adapters.PortfolioHandler{Service: portfolioService}

	r := chi.NewRouter()
	r.Get("/api/portfolio/wallet", portfolioHandler.GetWallet)
	r.Get("/api/portfolio/positions", portfolioHandler.FetchPositions)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("{\"message\": \"Portfolio service OK\"}"))
		if err != nil {
			log.Errorf("Health check response error: %v", err)
		}
	})

	log.Println("Starting Portfolio Service on port " + config.Port)
	if err := http.ListenAndServe(":"+config.Port, r); err != nil {
		log.Fatalf("Error when starting service : %v", err)
	}
}

func initDbConnection() (*dao_adapters.SQLWalletRepository, *dao_adapters.SQLPositionRepository) {
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
	return &dao_adapters.SQLWalletRepository{DB: db}, &dao_adapters.SQLPositionRepository{DB: db}
}
