package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luminosita/change-me/internal/core/constants"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	startupTime time.Time
	version     string
}

// NewHealthHandler creates a new health check handler.
func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{
		startupTime: time.Now(),
		version:     version,
	}
}

// HealthCheckResponse represents health check response schema.
type HealthCheckResponse struct {
	Status        string  `json:"status" example:"healthy"`
	Version       string  `json:"version" example:"0.1.0"`
	UptimeSeconds float64 `json:"uptime_seconds" example:"123.45"`
	Timestamp     string  `json:"timestamp" example:"2024-01-15T10:30:00Z"`
}

// Check handles GET /health endpoint.
//
// @Summary Health check endpoint
// @Description Returns application health status, version, uptime, and timestamp
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} HealthCheckResponse
// @Router /health [get]
func (h *HealthHandler) Check(c *gin.Context) {
	currentTime := time.Now()
	uptime := currentTime.Sub(h.startupTime).Seconds()

	response := HealthCheckResponse{
		Status:        constants.HealthStatusHealthy,
		Version:       h.version,
		UptimeSeconds: uptime,
		Timestamp:     currentTime.UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}
