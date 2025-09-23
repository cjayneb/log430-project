package main

import (
	"brokerx/adapters"
	"brokerx/core"
	"database/sql"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-sql-driver/mysql"
)

var config Config = Config{}

func main() {
	config.LoadConfig()
	
    db, e := sql.Open("mysql", config.DBUrl)
	if err := db.Ping(); err != nil || e != nil {
		log.Fatalf("Db error : %s | %s", e, err)
		return
	}
	userRepo := &adapters.SQLUserRepository{DB: db}

	authService := &core.AuthService{Repo: userRepo}
	authHandler := &adapters.AuthHandler{Service: *authService}

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Post("/login", authHandler.Login)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.ListenAndServe(":"+config.Port, router)
}