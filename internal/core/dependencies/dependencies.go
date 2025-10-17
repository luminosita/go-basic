package dependencies

import (
	"net/http"
	"time"

	"github.com/luminosita/change-me/internal/config"
	"github.com/luminosita/change-me/pkg/logger"
)

// Container holds all application dependencies.
// Acts as a dependency injection container initialized at startup.
type Container struct {
	Config     *config.Config
	Logger     *logger.Logger
	HTTPClient *http.Client
}

// NewContainer creates a new dependency injection container.
// Initializes all application-scoped dependencies.
//
// Parameters:
//   - cfg: Application configuration
//   - log: Structured logger
//
// Returns:
//   - *Container: Initialized dependency container
func NewContainer(cfg *config.Config, log *logger.Logger) *Container {
	// Create shared HTTP client with connection pooling
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &Container{
		Config:     cfg,
		Logger:     log,
		HTTPClient: httpClient,
	}
}

// Close cleans up resources held by the container.
// Should be called during application shutdown.
func (c *Container) Close() error {
	// Close HTTP client connections
	c.HTTPClient.CloseIdleConnections()

	// Sync logger (flush buffered entries)
	if err := c.Logger.Sync(); err != nil {
		return err
	}

	return nil
}
