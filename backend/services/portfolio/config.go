package main

import (
	"github.com/caarlos0/env"
)

type Config struct {
	Port string `env:"APP_PORT" envDefault:"8080"`

	DBUrl string `env:"DATABASE_URL" envDefault:"root:root@tcp(127.0.0.1:3306)/brokerx?parseTime=true"`

	KafkaHost							string `env:"KAFKA_HOST" envDefault:"127.0.0.1:9092"`
	KafkaGroupId						string `env:"KAFKA_GROUP_ID" envDefault:"group3"`
}

func (config *Config) LoadConfig() error {
	if err := env.Parse(config); err != nil {
		return err
	}
	return nil
}
