package main

import (
	"github.com/caarlos0/env"
)

type Config struct {
	Port string `env:"APP_PORT" envDefault:"8080"`

	ResourcePath string `env:"RESOURCES_PATH" envDefault:"resources/"`
}

func (config *Config) LoadConfig() error {
	if err := env.Parse(config); err != nil {
		return err
	}
	return nil
}
