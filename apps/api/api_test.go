package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"

	"github.com/stretchr/testify/assert"
)

func TestRoutes(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Ping endpoint",
			method:         "GET",
			path:           "/api/v1/ping",
			expectedStatus: 200,
			expectedBody:   "pong",
		},
		{
			name:           "Test Healhtz",
			method:         "GET",
			path:           "/api/v1/healthz",
			expectedStatus: 200,
			expectedBody:   "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != "" {
				var response struct {
					Message string `json:"message"`
				}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, tt.expectedBody, response.Message)
			}
		})
	}
}
