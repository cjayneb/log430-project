package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_main(t *testing.T) {
	tests := []struct {
		name string // description of this test case
	}{
		{name: "Main runs without errors"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("APP_PORT", "8081")
			defer os.Clearenv()
			router := run() // returns http.Handler

			req := httptest.NewRequest("GET", "/health", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			body, err := io.ReadAll(rec.Body)
			assert.Nil(t, err)
			assert.Contains(t, string(body), "OK")
		})
	}
}
