package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingEndpoint(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "pong", response["message"])
}

func TestHealthzEndpoint(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/healthz", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "ok", response["status"])
}

func TestWorkEndpoint(t *testing.T) {
	// Reset config for predictable test
	config.LatencyMS = 0
	config.FailRate = 0
	config.MemoryMB = 0

	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/work", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "completed", response["status"])
}

func TestWorkEndpointWithLatency(t *testing.T) {
	config.LatencyMS = 50 // 50ms delay
	config.FailRate = 0
	config.MemoryMB = 0

	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/work", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Should take at least 50ms
	durationMs := response["duration_ms"].(float64)
	assert.GreaterOrEqual(t, durationMs, float64(50))
}

func TestWorkEndpointFailure(t *testing.T) {
	config.LatencyMS = 0
	config.FailRate = 1.0 // Always fail
	config.MemoryMB = 0

	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/work", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "simulated failure", response["error"])
}

func TestEchoEndpoint(t *testing.T) {
	router := setupRouter()

	body := bytes.NewBufferString(`{"hello": "world", "count": 42}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/echo", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	echo := response["echo"].(map[string]interface{})
	assert.Equal(t, "world", echo["hello"])
	assert.Equal(t, float64(42), echo["count"])
}

func TestEchoEndpointInvalidJSON(t *testing.T) {
	router := setupRouter()

	body := bytes.NewBufferString(`not valid json`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/echo", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}
