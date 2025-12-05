package main

import (
	client_adapters "brokerx/portfolio-service/adapters/clients"
	dao_adapters "brokerx/portfolio-service/adapters/dao"
	handler_adapters "brokerx/portfolio-service/adapters/handlers"
	"brokerx/portfolio-service/core"
	"brokerx/portfolio-service/util"
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/traceid"
)

const TraceIDHeader = "X-Trace-Id"
type contextKey string
const TraceIdCtxKey contextKey = "traceId"

var config Config = Config{}

func main() {
	InitLogger("portfolio-service")

	if err := config.LoadConfig(); err != nil {
		slog.Error("Config error", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting Portfolio Service", "port", config.Port)
	router := run()
	if err := http.ListenAndServe(":"+config.Port, router); err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
}

func run() http.Handler {
	walletRepo, positionsRepo, outboxRepo, tm := initDbConnection()

	portfolioService := &core.PortfolioServiceImpl{
		WalletRepo:    walletRepo,
		PositionsRepo: positionsRepo,
		Tm: &tm,
	}
	portfolioHandler := handler_adapters.PortfolioHandler{Service: portfolioService}

	eventProducer := client_adapters.NewKafkaEventProducer("kafka:9092")
	orderMatchedHandler := handler_adapters.OrderMatchedHandler{
		Producer: eventProducer,
		Tm: &tm,
	}
	matchingEventConsumer := handler_adapters.NewKafkaMatchingEventConsumer(config.KafkaHost, config.KafkaMatchGroupId, orderMatchedHandler)
	go matchingEventConsumer.Start("PortfolioEvents")

	orderCompliantHandler := handler_adapters.OrderCompliantHandler{Producer: eventProducer, Tm: &tm}
	orderEventConsumer := handler_adapters.NewKafkaOrderEventConsumer(config.KafkaHost, config.KafkaOrderGroupId, orderCompliantHandler)
	go orderEventConsumer.Start("PortfolioEvents")

	outboxDispatcher := core.NewOutboxDispatcher(outboxRepo, eventProducer, 500 * time.Millisecond, 100, 10, 1 * time.Second)
	go outboxDispatcher.Start()


	r := chi.NewRouter()
	r.Use(httplog.RequestLogger(logger()))
	r.Use(middleware.Recoverer)
	r.Use(TraceMiddleware)

	r.Get("/api/portfolio/wallet", portfolioHandler.GetWallet)
	r.Patch("/api/portfolio/wallet/fund", portfolioHandler.FundWallet)
	r.Get("/api/portfolio/positions", portfolioHandler.FetchPositionsForSymbol)
	r.Get("/api/portfolio/positions/user", portfolioHandler.FetchPositionsForUser)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		log := util.FromContext(r.Context())
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("{\"message\": \"Portfolio service OK\"}"))
		if err != nil {
			log.Error("Health check response error", "error", err)
		}
	})

	return r
}

func initDbConnection() (*dao_adapters.SQLWalletRepository, *dao_adapters.SQLPositionRepository, *dao_adapters.SQLOutboxRepository, dao_adapters.SQLTransactionManager) {
	db, err := sql.Open("mysql", config.DBUrl)
	if err != nil {
		slog.Error("Db open error", "error", err)
		os.Exit(1)
	}

	db.SetMaxOpenConns(35)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Minute * 5)
	db.SetConnMaxIdleTime(time.Minute * 1)

	if err := db.Ping(); err != nil {
		slog.Warn("Db ping error", "error", err)
	}
	return &dao_adapters.SQLWalletRepository{DB: db}, &dao_adapters.SQLPositionRepository{DB: db}, &dao_adapters.SQLOutboxRepository{DB: db}, dao_adapters.SQLTransactionManager{DB: db}
}

func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		traceID := r.Header.Get(TraceIDHeader)
		if traceID == "" {
			traceID = uuid.New().String()
		}

		ctx = traceid.NewContext(ctx)
		ctx = context.WithValue(ctx, TraceIdCtxKey, traceID)

		reqLogger := slog.Default().With("traceId", traceID)
		ctx = util.WithLogger(ctx, reqLogger)

		httplog.LogEntrySetField(ctx, "traceId", slog.StringValue(traceID))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}


func InitLogger(service string) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	})

	logger := slog.New(handler).With("service", service)

	slog.SetDefault(logger)
}

func logger() *httplog.Logger {
	return httplog.NewLogger("portfolio-service", httplog.Options{
		LogLevel:         slog.LevelInfo,
		RequestHeaders:   false,
		ResponseHeaders:  false,
		JSON:             false,
		Concise:          true,
		MessageFieldName: "message",
		LevelFieldName:   "severity",
		TimeFieldFormat:  time.RFC3339,
	})
}
