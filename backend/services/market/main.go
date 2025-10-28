package main

import (
	handler_adapters "brokerx/market-data-service/adapters/handlers"
	"brokerx/market-data-service/core"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi/v5"
)

var config Config = Config{}

func main() {
	log.Println("Starting Market data Service on port " + config.Port)
	router := run()
	if err := http.ListenAndServe(":"+config.Port, router); err != nil {
		log.Fatalf("Server error : %s", err)
	}
}

func run() http.Handler {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Config error : %s", err)
	}

	marketDataService := core.NewMarketDataServiceImpl(config.ResourcePath)
	marketDataHandler := handler_adapters.MarketDataHandler{Service: marketDataService}

	r := chi.NewRouter()
	r.Get("/api/market/stock/price", marketDataHandler.GetStockPrice)
	r.Get("/api/market/stock/instrument", marketDataHandler.GetInstrument)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("{\"message\": \"Market data service OK\"}"))
		if err != nil {
			log.Errorf("Health check response error: %v", err)
		}
	})
	return r
}
