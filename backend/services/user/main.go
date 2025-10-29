package main

import (
	dao_adapters "brokerx/user-service/adapters/dao"
	handler_adapters "brokerx/user-service/adapters/handlers"
	"brokerx/user-service/core"
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

	userRepo := initDbConnection()
	authService := &core.AuthServiceImpl{
		Repo: userRepo,
		PasswordAllowedRetries: config.PasswordAllowedRetries,
		PasswordLockDurationMinutes: config.PasswordLockDurationMinutes,
	}
	authHandler := &handler_adapters.AuthHandler{
		Service:   authService,
		JWTSecret: []byte(config.JWTSecret),
	}

	r := chi.NewRouter()
	r.Post("/api/user/register", authHandler.Register)
	r.Post("/api/user/auth/login", authHandler.Login)
	r.Get("/api/user/auth/verify", authHandler.VerifyToken)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("{\"message\": \"User service OK\"}"))
		if err != nil {
			log.Errorf("Health check response error: %v", err)
		}
	})

	log.Println("Starting User Service on port " + config.Port)
	if err := http.ListenAndServe(":"+config.Port, r); err != nil {
		log.Fatalf("Error when starting service : %v", err)
	}
}

func initDbConnection() *dao_adapters.SQLUserRepository {
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
	return &dao_adapters.SQLUserRepository{DB: db}
}
