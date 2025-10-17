//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luminosita/change-me/internal/config"
	"github.com/luminosita/change-me/internal/core/dependencies"
	httpserver "github.com/luminosita/change-me/internal/interfaces/http"
	"github.com/luminosita/change-me/internal/interfaces/http/handlers"
	"github.com/luminosita/change-me/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ====================
// Test Helpers
// ====================

// setupTestServer creates a test HTTP server with real dependencies
func setupTestServer(t *testing.T) (*httpserver.Server, *dependencies.Container) {
	t.Helper()

	// Create test configuration manually
	cfg := &config.Config{
		AppName:    "Test Server",
		AppVersion: "0.1.0",
		Debug:      true,
		Host:       "127.0.0.1",
		Port:       0, // Random port
		LogLevel:   "INFO",
		LogFormat:  "json",
	}

	// Create logger
	log, err := logger.New(logger.Config{
		Level:  cfg.LogLevel,
		Format: cfg.LogFormat,
	})
	require.NoError(t, err)

	// Initialize container manually (not using Wire for tests)
	container := dependencies.NewContainer(cfg, log)
	t.Cleanup(func() {
		_ = container.Close()
	})

	// Create server
	server := httpserver.New(container)
	return server, container
}

// HealthCheckResponse matches the response structure
type HealthCheckResponse struct {
	Status        string  `json:"status"`
	Version       string  `json:"version"`
	UptimeSeconds float64 `json:"uptime_seconds"`
	Timestamp     string  `json:"timestamp"`
}

// ====================
// Synchronous Endpoint Tests
// ====================

func TestHealthEndpoint_ReturnsOKStatus(t *testing.T) {
	// Arrange
	server, _ := setupTestServer(t)
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Act
	server.Router().ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response HealthCheckResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response.Status)
	assert.NotEmpty(t, response.Version)
	assert.GreaterOrEqual(t, response.UptimeSeconds, 0.0)
	assert.NotEmpty(t, response.Timestamp)
}

func TestHealthEndpoint_ReturnsCorrectContentType(t *testing.T) {
	// Arrange
	server, _ := setupTestServer(t)
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Act
	server.Router().ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestHealthEndpoint_ReturnsConsistentResponseOnMultipleCalls(t *testing.T) {
	// Arrange
	server, _ := setupTestServer(t)

	// Act - make multiple requests
	responses := make([]HealthCheckResponse, 3)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		server.Router().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		err := json.Unmarshal(w.Body.Bytes(), &responses[i])
		require.NoError(t, err)
	}

	// Assert - status and version should be identical
	for i := 1; i < 3; i++ {
		assert.Equal(t, responses[0].Status, responses[i].Status)
		assert.Equal(t, responses[0].Version, responses[i].Version)
	}
}

// ====================
// Concurrent Request Tests
// ====================

func TestHealthEndpoint_HandlesConcurrentRequests(t *testing.T) {
	// Arrange
	server, _ := setupTestServer(t)
	const concurrentRequests = 10

	// Act - make concurrent requests
	var wg sync.WaitGroup
	results := make([]int, concurrentRequests)
	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			server.Router().ServeHTTP(w, req)
			results[index] = w.Code
		}(i)
	}
	wg.Wait()

	// Assert - all requests should succeed
	for i, code := range results {
		assert.Equal(t, http.StatusOK, code, "Request %d failed", i)
	}
}

// ====================
// Error Handling Tests
// ====================

func TestNonExistentEndpoint_Returns404(t *testing.T) {
	// Arrange
	server, _ := setupTestServer(t)
	req := httptest.NewRequest("GET", "/nonexistent-endpoint", nil)
	w := httptest.NewRecorder()

	// Act
	server.Router().ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	// Gin returns text/plain for 404 by default (without custom handler)
	contentType := w.Header().Get("Content-Type")
	assert.True(t,
		contentType == "text/plain; charset=utf-8" ||
			contentType == "text/plain" ||
			contentType == "",
		"Expected text/plain (with or without charset) or empty content-type for 404, got: %s", contentType)
}

func TestUnsupportedHTTPMethod_Returns404Or405(t *testing.T) {
	// Arrange - health endpoint only supports GET
	server, _ := setupTestServer(t)
	req := httptest.NewRequest("POST", "/health", nil)
	w := httptest.NewRecorder()

	// Act
	server.Router().ServeHTTP(w, req)

	// Assert - Gin returns 404 by default for unmatched routes (even wrong methods)
	// To get 405, you need to configure HandleMethodNotAllowed
	// See: https://gin-gonic.com/docs/examples/custom-http-config/
	assert.Equal(t, http.StatusNotFound, w.Code,
		"Gin returns 404 for unmatched method by default (without HandleMethodNotAllowed=true)")
}

// ====================
// Response Validation Tests
// ====================

func TestHealthEndpoint_ValidatesResponseSchema(t *testing.T) {
	// Arrange
	server, _ := setupTestServer(t)
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Act
	server.Router().ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response HealthCheckResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Validate field types and constraints
	assert.IsType(t, "", response.Status)
	assert.IsType(t, "", response.Version)
	assert.IsType(t, 0.0, response.UptimeSeconds)
	assert.IsType(t, "", response.Timestamp)

	// Validate field values
	validStatuses := []string{"healthy", "degraded", "unhealthy"}
	assert.Contains(t, validStatuses, response.Status)
	assert.Greater(t, len(response.Version), 0)
	assert.GreaterOrEqual(t, response.UptimeSeconds, 0.0)
	assert.Contains(t, response.Timestamp, "T") // ISO 8601 format
}

func TestHealthEndpoint_TimestampIsRFC3339Format(t *testing.T) {
	// Arrange
	server, _ := setupTestServer(t)
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Act
	server.Router().ServeHTTP(w, req)

	// Assert
	var response HealthCheckResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify RFC3339 format (ISO 8601)
	_, err = time.Parse(time.RFC3339, response.Timestamp)
	assert.NoError(t, err, "timestamp should be in RFC3339 format")
}

func TestHealthEndpoint_UptimeIncreases(t *testing.T) {
	// Arrange
	server, _ := setupTestServer(t)

	// Act - first request
	req1 := httptest.NewRequest("GET", "/health", nil)
	w1 := httptest.NewRecorder()
	server.Router().ServeHTTP(w1, req1)

	var response1 HealthCheckResponse
	err := json.Unmarshal(w1.Body.Bytes(), &response1)
	require.NoError(t, err)

	// Wait a short time
	time.Sleep(100 * time.Millisecond)

	// Act - second request
	req2 := httptest.NewRequest("GET", "/health", nil)
	w2 := httptest.NewRecorder()
	server.Router().ServeHTTP(w2, req2)

	var response2 HealthCheckResponse
	err = json.Unmarshal(w2.Body.Bytes(), &response2)
	require.NoError(t, err)

	// Assert - uptime should have increased
	assert.Greater(t, response2.UptimeSeconds, response1.UptimeSeconds)
}

// ====================
// Integration with Real Server
// ====================

func TestRealHTTPServer_HealthEndpoint(t *testing.T) {
	// Arrange - setup test configuration with random port
	cfg := &config.Config{
		AppName:    "Test Server",
		AppVersion: "0.1.0",
		Debug:      true,
		Host:       "127.0.0.1",
		Port:       0, // OS will assign random port
		LogLevel:   "INFO",
		LogFormat:  "json",
	}

	log, err := logger.New(logger.Config{
		Level:  cfg.LogLevel,
		Format: cfg.LogFormat,
	})
	require.NoError(t, err)

	container := dependencies.NewContainer(cfg, log)
	defer container.Close()

	// Create server
	gin.SetMode(gin.TestMode)
	router := gin.New()
	healthHandler := handlers.NewHealthHandler(cfg.AppVersion)
	router.GET("/health", healthHandler.Check)

	// Start test server
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// Act - make real HTTP request
	resp, err := http.Get(fmt.Sprintf("%s/health", testServer.URL))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")

	var response HealthCheckResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response.Status)
}
