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

	log.Println("Starting Market data Service on port " + config.Port)
	http.ListenAndServe(":"+config.Port, r)
}
