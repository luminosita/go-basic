package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luminosita/change-me/internal/core/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthCheck_Returns200OK(t *testing.T) {
	router, _ := setupHealthTest()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthCheck_ReturnsJSON(t *testing.T) {
	router, _ := setupHealthTest()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var data map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &data)
	assert.NoError(t, err)
}

func TestHealthCheck_ResponseSchema(t *testing.T) {
	router, _ := setupHealthTest()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var data HealthCheckResponse
	err := json.Unmarshal(w.Body.Bytes(), &data)
	require.NoError(t, err)

	// Verify all required fields present
	assert.NotEmpty(t, data.Status)
	assert.NotEmpty(t, data.Version)
	assert.GreaterOrEqual(t, data.UptimeSeconds, 0.0)
	assert.NotEmpty(t, data.Timestamp)
}

func TestHealthCheck_StatusValue(t *testing.T) {
	router, _ := setupHealthTest()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var data HealthCheckResponse
	err := json.Unmarshal(w.Body.Bytes(), &data)
	require.NoError(t, err)

	assert.Equal(t, constants.HealthStatusHealthy, data.Status)
}

func TestHealthCheck_VersionFormat(t *testing.T) {
	router, _ := setupHealthTest()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var data HealthCheckResponse
	err := json.Unmarshal(w.Body.Bytes(), &data)
	require.NoError(t, err)

	assert.NotEmpty(t, data.Version)
	// Verify version follows semver format (basic check)
	assert.Equal(t, "0.1.0", data.Version)
}

func TestHealthCheck_UptimeIsPositive(t *testing.T) {
	router, _ := setupHealthTest()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var data HealthCheckResponse
	err := json.Unmarshal(w.Body.Bytes(), &data)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, data.UptimeSeconds, 0.0)
}

func TestHealthCheck_UptimeIncreases(t *testing.T) {
	router, _ := setupHealthTest()

	// First request
	req1 := httptest.NewRequest("GET", "/health", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	var data1 HealthCheckResponse
	err := json.Unmarshal(w1.Body.Bytes(), &data1)
	require.NoError(t, err)

	// Wait a short time
	time.Sleep(100 * time.Millisecond)

	// Second request
	req2 := httptest.NewRequest("GET", "/health", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	var data2 HealthCheckResponse
	err = json.Unmarshal(w2.Body.Bytes(), &data2)
	require.NoError(t, err)

	// Uptime should have increased
	assert.Greater(t, data2.UptimeSeconds, data1.UptimeSeconds)
}

func TestHealthCheck_TimestampFormat(t *testing.T) {
	router, _ := setupHealthTest()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var data HealthCheckResponse
	err := json.Unmarshal(w.Body.Bytes(), &data)
	require.NoError(t, err)

	// Verify RFC3339 format (ISO 8601)
	_, err = time.Parse(time.RFC3339, data.Timestamp)
	assert.NoError(t, err, "timestamp should be in RFC3339 format")
}

func TestHealthCheck_MultipleCallsIdempotent(t *testing.T) {
	router, _ := setupHealthTest()

	var responses []HealthCheckResponse
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var data HealthCheckResponse
		err := json.Unmarshal(w.Body.Bytes(), &data)
		require.NoError(t, err)

		responses = append(responses, data)
	}

	// All responses should have the same status and version
	for _, data := range responses {
		assert.Equal(t, constants.HealthStatusHealthy, data.Status)
		assert.Equal(t, responses[0].Version, data.Version)
	}
}

// setupHealthTest creates a test Gin router with health handler
func setupHealthTest() (*gin.Engine, *HealthHandler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHealthHandler("0.1.0")
	router.GET("/health", handler.Check)
	return router, handler
}
