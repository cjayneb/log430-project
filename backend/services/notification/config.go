package main

import (
	"github.com/caarlos0/env"
)

type Config struct {
	ApiGatewayBaseUrl string `env:"API_GATEWAY_BASE_URL" envDefault:"http://localhost"`
	Port              string `env:"APP_PORT" envDefault:"8080"`

	DBUrl     string `env:"DATABASE_URL" envDefault:"root:root@tcp(127.0.0.1:3306)/brokerx?parseTime=true"`
	RedisAddr string `env:"REDIS_ADDR" envDefault:"127.0.0.1:6379"`

	KafkaHost    string `env:"KAFKA_HOST" envDefault:"127.0.0.1:9092"`
	KafkaGroupId string `env:"KAFKA_GROUP_ID" envDefault:"notif-group"`
}

func (config *Config) LoadConfig() error {
	if err := env.Parse(config); err != nil {
		return err
	}
	return nil
}
