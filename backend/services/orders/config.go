package main

import (
	"github.com/caarlos0/env"
)

type Config struct {
	Port                        		string `env:"APP_PORT" envDefault:"8080"`
	ApiGatewayBaseUrl		   			string `env:"API_GATEWAY_BASE_URL" envDefault:"http://localhost"`	

	DBUrl                       		string `env:"DATABASE_URL" envDefault:"root:root@tcp(127.0.0.1:3306)/brokerx?parseTime=true"`
	RedisAddr							string `env:"REDIS_ADDR" envDefault:"127.0.0.1:6379"`

	NumberOfGoRoutines 					int `env:"NUMBER_OF_GO_ROUTINES" envDefault:"8"`

	DirtyOrderSyncIntervalInSeconds 	int `env:"DIRTY_ORDER_SYNC_INTERVAL_SECONDS" envDefault:"1"`
	DirtyOrderSyncBatchSize 			int `env:"DIRTY_ORDER_SYNC_BATCH_SIZE" envDefault:"100"`
	OrdersExecutionsPersistIntervalInMs int `env:"ORDERS_EXECS_PERSIST_INTERVAL_MS" envDefault:"300"`
	OrdersPersistBatchSize 				int `env:"ORDERS_PERSIST_BATCH_SIZE" envDefault:"100"`
	ExecutionsPersistBatchSize 			int `env:"EXECS_PERSIST_BATCH_SIZE" envDefault:"200"`
}

func (config *Config) LoadConfig() error {
	if err := env.Parse(config); err != nil {
		return err
	}
	return nil
}
