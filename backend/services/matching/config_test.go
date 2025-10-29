package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigDefaults(t *testing.T) {
	os.Clearenv() // remove all env vars

	cfg := Config{}
	err := cfg.LoadConfig()

	assert.Nil(t, err)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, 8, cfg.NumberOfGoRoutines)
	assert.Equal(t, "127.0.0.1:6379", cfg.RedisAddr)
}

func TestLoadConfigCustomValues(t *testing.T) {
	os.Setenv("APP_PORT", "9999")
	os.Setenv("NUMBER_OF_GO_ROUTINES", "10")
	defer os.Clearenv()

	cfg := Config{}
	err := cfg.LoadConfig()

	assert.Nil(t, err)
	assert.Equal(t, "9999", cfg.Port)
	assert.Equal(t, 10, cfg.NumberOfGoRoutines)
}

func TestLoadConfigError(t *testing.T) {
	os.Setenv("NUMBER_OF_GO_ROUTINES", "not-an-int")
	defer os.Clearenv()

	cfg := Config{}
	err := cfg.LoadConfig()

	assert.NotNil(t, err)
}
