package main

import (
	"github.com/caarlos0/env"
)

type Config struct {
	Port string `env:"APP_PORT" envDefault:"8080"`

	NumberOfGoRoutines 					int `env:"NUMBER_OF_GO_ROUTINES" envDefault:"8"`

	RedisAddr string `env:"REDIS_ADDR" envDefault:"127.0.0.1:6379"`

	KafkaHost							string `env:"KAFKA_HOST" envDefault:"127.0.0.1:9092"`
	KafkaGroupId						string `env:"KAFKA_GROUP_ID" envDefault:"group2"`
}

func (config *Config) LoadConfig() error {
	if err := env.Parse(config); err != nil {
		return err
	}
	return nil
}
