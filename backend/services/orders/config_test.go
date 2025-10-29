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
	assert.Equal(t, "root:root@tcp(127.0.0.1:3306)/brokerx?parseTime=true", cfg.DBUrl)
}

func TestLoadConfigCustomValues(t *testing.T) {
	os.Setenv("APP_PORT", "9999")
	os.Setenv("PASSWORD_ALLOWED_RETRIES", "10")
	defer os.Clearenv()

	cfg := Config{}
	err := cfg.LoadConfig()

	assert.Nil(t, err)
	assert.Equal(t, "9999", cfg.Port)
}

func TestLoadConfigError(t *testing.T) {
	os.Setenv("DIRTY_ORDER_SYNC_INTERVAL_SECONDS", "not-an-int")
	defer os.Clearenv()

	cfg := Config{}
	err := cfg.LoadConfig()

	assert.NotNil(t, err)
}
