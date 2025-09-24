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
	
	userRepo:= initDbConnection(&config)
	authService := &core.AuthService{Repo: userRepo}
	authHandler := &adapters.AuthHandler{Service: *authService}

	router := initRouter(authHandler)
	http.ListenAndServe(":"+config.Port, router)
}

func initDbConnection(config *Config) (*adapters.SQLUserRepository) {
	db, e := sql.Open("mysql", config.DBUrl)
	if err := db.Ping(); err != nil || e != nil {
		log.Fatalf("Db error : %s | %s", e, err)
	}
	return &adapters.SQLUserRepository{DB: db}
}

func initRouter(authHandler *adapters.AuthHandler) (*chi.Mux) {
	router := chi.NewRouter()
    router.Use(middleware.Logger)

	// Public static assets
    fs := http.StripPrefix("/static/", http.FileServer(http.Dir("./frontend/static")))
    router.Handle("/static/*", fs)
    router.Get("/login", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "./frontend/login.html")
    })

    // Protected routes
    router.Group(func(r chi.Router) {
        r.Use(authHandler.Middleware)
		r.Use(middleware.Logger)
        r.Get("/", func(w http.ResponseWriter, r *http.Request) {
            http.ServeFile(w, r, "./frontend/index.html")
        })
    })

    // API
    router.Post("/auth/login", authHandler.Login)

    // Health
    router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })

    return router
}