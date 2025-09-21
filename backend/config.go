package main

import (
	"log"

	"github.com/caarlos0/env"
)

type Config struct {
	Port  string `env:"APP_PORT" envDefault:"8080"`
	DBUrl string `env:"DATABASE_URL" envDefault:"root:root@tcp(127.0.0.1:3306)/brokerx"`
}

func (config *Config) LoadConfig() {
	if err := env.Parse(config); err != nil {
		log.Fatalf("%+v", err)
	}
}