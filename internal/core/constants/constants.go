package constants

// Application metadata
const (
	AppName        = "CHANGE_ME"
	AppVersion     = "0.1.0"
	AppDescription = "Go HTTP server with health check, logging, and DI"
)

// API configuration
const (
	APIPrefix = "/api/v1"
	DocsURL   = "/docs"
	RedocURL  = "/redoc"
)

// Health check status values
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusDegraded  = "degraded"
	HealthStatusUnhealthy = "unhealthy"
)

// CORS configuration (development)
var (
	CORSAllowOrigins = []string{
		"http://localhost:3000",
		"http://localhost:8000",
		"http://localhost:8080",
	}
	CORSAllowMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	CORSAllowHeaders = []string{"*"}
)

// Logging
const (
	LogFormatJSON = "json"
	LogFormatText = "text"
)
