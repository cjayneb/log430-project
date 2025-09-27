package main

import (
	"brokerx/adapters"
	"brokerx/core"
	"database/sql"
	"html/template"
	"net/http"

	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-sql-driver/mysql"
)

var config Config = Config{}
var templates *template.Template

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

    userRepo, orderRepo := initDbConnection()
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

func initDbConnection() (*adapters.SQLUserRepository, *adapters.SQLOrderRepository) {
	db, e := sql.Open("mysql", config.DBUrl)
	if err := db.Ping(); err != nil || e != nil {
		log.Warnf("Db error : %s | %s", e, err)
	}
	return &adapters.SQLUserRepository{DB: db}, &adapters.SQLOrderRepository{DB: db}
}

func initRouter(authHandler *adapters.AuthHandler, orderHandler *adapters.OrderHandler) (*chi.Mux) {
	router := chi.NewRouter()
    router.Use(middleware.Logger)
    router.Use(noCacheMiddleware)

	// Public static assets
    fs := http.StripPrefix("/static/", http.FileServer(http.Dir(config.FrontendPath+"/static")))
    router.Handle("/static/*", fs)
    router.Get("/login", func(w http.ResponseWriter, r *http.Request) {
        renderTemplate(w, "login.html", nil)
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
            userEmail := r.Context().Value(adapters.USER_EMAIL_KEY).(string)
            renderTemplate(w, "index.html", map[string]string{"Email": userEmail})
        })

        r.Get("/order", func(w http.ResponseWriter, r *http.Request) {
            userEmail := r.Context().Value(adapters.USER_EMAIL_KEY).(string)
            renderTemplate(w, "order.html", map[string]string{"Email": userEmail})
        })

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

func renderTemplate(w http.ResponseWriter, name string, data any) {
    tpl, err := template.ParseFiles(config.FrontendPath+"/templates/base.html", config.FrontendPath+"/templates/"+name)
    if err != nil {
        http.Error(w, "Template parse error: "+err.Error(), http.StatusInternalServerError)
        return
    }

	err = tpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
	}
}