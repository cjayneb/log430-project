package main

import (
	"github.com/caarlos0/env"
)

type Config struct {
	Port                        		string `env:"APP_PORT" envDefault:"8080"`
	ApiGatewayBaseUrl		   			string `env:"API_GATEWAY_BASE_URL" envDefault:"http://localhost"`
	MarketDataServiceBaseUrl			string `env:"MARKET_DATA_SERVICE_BASE_URL" envDefault:"http://market-data-service:8080"`
	MatchingServiceBaseUrl				string `env:"MATCHING_SERVICE_BASE_URL" envDefault:"http://matching-service:8080"`
	PortfolioServiceBaseUrl				string `env:"PORTFOLIO_SERVICE_BASE_URL" envDefault:"http://portfolio-service:8080"`

	DBUrl                       		string `env:"DATABASE_URL" envDefault:"root:root@tcp(127.0.0.1:3306)/brokerx?parseTime=true"`
	RedisAddr							string `env:"REDIS_ADDR" envDefault:"127.0.0.1:6379"`

	KafkaHost							string `env:"KAFKA_HOST" envDefault:"127.0.0.1:9092"`
	KafkaGroupId						string `env:"KAFKA_GROUP_ID" envDefault:"group1"`

	DirtyOrderSyncIntervalInSeconds 	int `env:"DIRTY_ORDER_SYNC_INTERVAL_SECONDS" envDefault:"1"`
	DirtyOrderSyncBatchSize 			int `env:"DIRTY_ORDER_SYNC_BATCH_SIZE" envDefault:"1000"`
	OrdersExecutionsPersistIntervalInMs int `env:"ORDERS_EXECS_PERSIST_INTERVAL_MS" envDefault:"300"`
	OrdersPersistBatchSize 				int `env:"ORDERS_PERSIST_BATCH_SIZE" envDefault:"500"`
	ExecutionsPersistBatchSize 			int `env:"EXECS_PERSIST_BATCH_SIZE" envDefault:"1000"`
}

func (config *Config) LoadConfig() error {
	if err := env.Parse(config); err != nil {
		return err
	}
	return nil
}
