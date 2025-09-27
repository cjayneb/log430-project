package main

import (
	"brokerx/adapters"
	"brokerx/core"
	"database/sql"
	"net/http"

	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-sql-driver/mysql"
)

var config Config = Config{}

func main() {
    router := run()
	if err := http.ListenAndServe(":"+config.Port, router); err != nil {
    	log.Fatalf("Server error : %s", err)
	}
}

func run() http.Handler {
    if err := config.LoadConfig(); err != nil {
		log.Fatalf("Config error : %s", err)
	}

    userRepo, orderRepo := initDbConnection(&config)
    authService := &core.AuthService{
        Repo:                        userRepo,
        PasswordAllowedRetries:      config.PasswordAllowedRetries,
        PasswordLockDurationMinutes: config.PasswordLockDurationMinutes,
    }
    authHandler := &adapters.AuthHandler{
        Service:      authService,
        SessionStore: sessions.NewCookieStore([]byte("very-secret-key")),
        IsProduction: config.IsProduction,
    }

    orderService := &core.OrderService{Repo: orderRepo}
    orderHandler := &adapters.OrderHandler{Service: orderService}

    router := initRouter(authHandler, orderHandler)
    return router
}

func initDbConnection(config *Config) (*adapters.SQLUserRepository, *adapters.SQLOrderRepository) {
	db, e := sql.Open("mysql", config.DBUrl)
	if err := db.Ping(); err != nil || e != nil {
		log.Warnf("Db error : %s | %s", e, err)
	}
	return &adapters.SQLUserRepository{DB: db}, &adapters.SQLOrderRepository{DB: db}
}

func initRouter(authHandler *adapters.AuthHandler, orderHandler *adapters.OrderHandler) (*chi.Mux) {
	router := chi.NewRouter()
    router.Use(middleware.Logger)

	// Public static assets
    fs := http.StripPrefix("/static/", http.FileServer(http.Dir("./frontend/static")))
    noCacheHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Prevent browser from using cached version
        w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
        w.Header().Set("Pragma", "no-cache")
        w.Header().Set("Expires", "0")
        fs.ServeHTTP(w, r)
    })
    router.Handle("/static/*", noCacheHandler)
    router.Get("/login", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "./frontend/login.html")
    })

	// Public API routes
    router.Post("/auth/login", authHandler.Login)
    router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        _, err := w.Write([]byte("OK"))
		if err != nil {
			log.Errorf("Health check response error: %v", err)
		}
    })

    // Protected routes
    router.Group(func(r chi.Router) {
        r.Use(authHandler.Middleware)
		r.Use(middleware.Logger)
        r.Get("/", func(w http.ResponseWriter, r *http.Request) {
            http.ServeFile(w, r, "./frontend/index.html")
        })

        r.Get("/order", func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
            http.ServeFile(w, r, "./frontend/order.html")
        })

        r.Post("/order/place", orderHandler.PlaceOrder)
    })

    return router
}