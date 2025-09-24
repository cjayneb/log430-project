package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/caarlos0/env"
)

type Config struct {
	Port  string `env:"APP_PORT" envDefault:"8080"`
	DBUrl string `env:"DATABASE_URL" envDefault:"root:root@tcp(127.0.0.1:3306)/brokerx?parseTime=true"`
	PasswordAllowedRetries int	`env:"PASSWORD_ALLOWED_RETRIES" envDefault:"5"`
	PasswordLockDurationMinutes int `env:"PASSWORD_LOCK_DURATION_MINUTES" envDefault:"30"`
}

func (config *Config) LoadConfig() {
	if err := env.Parse(config); err != nil {
		log.Fatalf("Error loading config : %v", err)
	}
}