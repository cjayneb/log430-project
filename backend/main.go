package main

import (
	"brokerx/adapters"
	"brokerx/core"
	"context"
	"database/sql"
	"html/template"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var cancel = context.CancelFunc(nil)

var config Config = Config{}

func main() {
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	router := run()
	if err := http.ListenAndServe(":"+config.Port, router); err != nil {
		log.Fatalf("Server error : %s", err)
	}
}

func run() http.Handler {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Config error : %s", err)
	}

	userRepo, orderRepo, walletRepo, positionRepo, transactionManager := initDbConnection()
	orderBook := initRedisConnection()
	authService := &core.AuthService{
		Repo:                        userRepo,
		PasswordAllowedRetries:      config.PasswordAllowedRetries,
		PasswordLockDurationMinutes: config.PasswordLockDurationMinutes,
	}
	authHandler := &adapters.AuthHandler{
		Service:      authService,
		SessionStore: sessions.NewCookieStore([]byte(uuid.New().String())),
		IsProduction: config.IsProduction,
	}

	complianceService := &core.ComplianceService{WalletRepo: walletRepo, PositionRepo: positionRepo, MarketDataProvider: adapters.NewMarketDataProvider(config.ResourcePath)}
	matchingEngine := &core.MatchingEngine{TransactionManager: transactionManager, OrderBook: orderBook}
	orderService := &core.OrderService{Repo: orderRepo, ComplianceService: complianceService, OrderBook: orderBook, MatchingEngine: matchingEngine}
	orderHandler := &adapters.OrderHandler{Service: orderService, FrontendPath: config.FrontendPath}

	orderService.StartMatchingWorkers()
	core.StartDirtyOrderSync(ctx, 1*time.Second, 100, orderBook, transactionManager)

	router := initRouter(authHandler, orderHandler)
	return router
}

func initDbConnection() (*adapters.SQLUserRepository, *adapters.SQLOrderRepository, *adapters.SQLWalletRepository, *adapters.SQLPositionRepository, *adapters.SQLTransactionManager) {
	db, e := sql.Open("mysql", config.DBUrl)

	db.SetMaxOpenConns(60)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Minute * 5)
	db.SetConnMaxIdleTime(time.Minute * 1)

	if err := db.Ping(); err != nil || e != nil {
		log.Warnf("Db error : %s | %s", e, err)
	}
	return &adapters.SQLUserRepository{DB: db},
		&adapters.SQLOrderRepository{DB: db},
		&adapters.SQLWalletRepository{DB: db},
		&adapters.SQLPositionRepository{DB: db},
		&adapters.SQLTransactionManager{DB: db}
}

func initRedisConnection() (*adapters.RedisOrderBook) {
	client := redis.NewClient(&redis.Options{
		Addr: config.RedisAddr,
		Password: "",
		DB: 0,
	})

	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Warnf("Redis error : %v", err)
	}

	// TODO: Initialize RedisOrderBook with the client
	return &adapters.RedisOrderBook{Rdb: client}
}

func initRouter(authHandler *adapters.AuthHandler, orderHandler *adapters.OrderHandler) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(noCacheMiddleware)

	// Public static assets
	fs := http.StripPrefix("/static/", http.FileServer(http.Dir(config.FrontendPath+"/static")))
	router.Handle("/static/*", fs)
	router.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, r, "login.html", nil)
	})

	// Public API routes
	router.Post("/auth/login", authHandler.Login)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("{\"message\": \"OK\"}"))
		if err != nil {
			log.Errorf("Health check response error: %v", err)
		}
	})
	router.Handle("/metrics", promhttp.Handler())

	// Protected routes
	router.Group(func(r chi.Router) {
		r.Use(authHandler.Middleware)
		r.Use(middleware.Logger)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			renderTemplate(w, r, "index.html", nil)
		})

		r.Get("/order", orderHandler.GetOrders)
		r.Post("/order/place", orderHandler.PlaceOrder)
	})

	return router
}

func noCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}

func renderTemplate(w http.ResponseWriter, r *http.Request, name string, data map[string]string) {
	tpl, err := template.ParseFiles(config.FrontendPath+"/templates/base.html", config.FrontendPath+"/templates/"+name)
	if err != nil {
		http.Error(w, "Template parse error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if data == nil {
		data = make(map[string]string)
	}

	userEmail := r.Context().Value(adapters.USER_EMAIL_KEY)
	if userEmail != nil {
		data["Email"] = userEmail.(string)
	}

	err = tpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
	}
}
