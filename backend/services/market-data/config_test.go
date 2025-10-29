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
	assert.Equal(t, "resources/", cfg.ResourcePath)
}

func TestLoadConfigCustomValues(t *testing.T) {
	os.Setenv("APP_PORT", "9999")
	os.Setenv("RESOURCES_PATH", "/resources/custom")
	defer os.Clearenv()

	cfg := Config{}
	err := cfg.LoadConfig()

	assert.Nil(t, err)
	assert.Equal(t, "9999", cfg.Port)
	assert.Equal(t, "/resources/custom", cfg.ResourcePath)
}