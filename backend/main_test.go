package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerEndpoint(t *testing.T) {
	os.Setenv("APP_PORT", "8081")
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		os.Setenv("DATABASE_URL", "root:root@tcp(127.0.0.1:3307)/brokerx?parseTime=true")
	}
	defer os.Clearenv()
    router := run() // returns http.Handler

    req := httptest.NewRequest("GET", "/health", nil)
    w := httptest.NewRecorder()

    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    body, err := io.ReadAll(w.Body)
    assert.Nil(t, err)
    assert.Contains(t, string(body), "OK")
}

func TestCoverageJustification(t *testing.T) {
	t.Skip("Full server tested externally via Postman")
}
