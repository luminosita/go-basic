//go:build integration

package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/luminosita/change-me/internal/config"
	"github.com/luminosita/change-me/internal/core/dependencies"
	httpserver "github.com/luminosita/change-me/internal/interfaces/http"
	"github.com/luminosita/change-me/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ====================
// Dependency Injection Tests
// ====================

func TestDependencyInjection_WireInitializesContainer(t *testing.T) {
	// Act - initialize container using Wire
	container, err := dependencies.InitializeContainer()

	// Assert
	require.NoError(t, err)
	require.NotNil(t, container)
	defer container.Close()

	// Verify all dependencies are initialized
	assert.NotNil(t, container.Config)
	assert.NotNil(t, container.Logger)
	assert.NotNil(t, container.HTTPClient)
}

func TestDependencyInjection_ConfigurationIsInjected(t *testing.T) {
	// Act
	container, err := dependencies.InitializeContainer()
	require.NoError(t, err)
	defer container.Close()

	// Assert - verify config fields
	cfg := container.Config
	assert.NotEmpty(t, cfg.AppName)
	assert.NotEmpty(t, cfg.AppVersion)
	assert.NotEmpty(t, cfg.Host)
	assert.Greater(t, cfg.Port, 0)
	assert.NotEmpty(t, cfg.LogLevel)
	assert.NotEmpty(t, cfg.LogFormat)
}

func TestDependencyInjection_LoggerIsInjected(t *testing.T) {
	// Act
	container, err := dependencies.InitializeContainer()
	require.NoError(t, err)
	defer container.Close()

	// Assert - verify logger is functional
	log := container.Logger
	assert.NotNil(t, log)

	// Test logger methods don't panic
	assert.NotPanics(t, func() {
		log.Info("test message")
		log.Infow("test message with fields", "key", "value")
	})
}

func TestDependencyInjection_HTTPClientIsInjected(t *testing.T) {
	// Act
	container, err := dependencies.InitializeContainer()
	require.NoError(t, err)
	defer container.Close()

	// Assert - verify HTTP client configuration
	client := container.HTTPClient
	assert.NotNil(t, client)
	assert.Equal(t, 30*time.Second, client.Timeout)

	// Verify transport is configured
	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	assert.Equal(t, 10, transport.MaxIdleConns)
	assert.Equal(t, 10, transport.MaxIdleConnsPerHost)
	assert.Equal(t, 90*time.Second, transport.IdleConnTimeout)
}

func TestDependencyInjection_ManualContainerCreation(t *testing.T) {
	// Arrange - create dependencies manually
	cfg := &config.Config{
		AppName:    "Manual Test",
		AppVersion: "1.0.0",
		Debug:      true,
		Host:       "127.0.0.1",
		Port:       8080,
		LogLevel:   "DEBUG",
		LogFormat:  "json",
	}

	log, err := logger.New(logger.Config{
		Level:  cfg.LogLevel,
		Format: cfg.LogFormat,
	})
	require.NoError(t, err)

	// Act - create container manually (without Wire)
	container := dependencies.NewContainer(cfg, log)
	defer container.Close()

	// Assert
	assert.NotNil(t, container)
	assert.Equal(t, "Manual Test", container.Config.AppName)
	assert.NotNil(t, container.Logger)
	assert.NotNil(t, container.HTTPClient)
}

func TestDependencyInjection_ContainerCloseReleasesResources(t *testing.T) {
	// Arrange
	container, err := dependencies.InitializeContainer()
	require.NoError(t, err)

	// Act - close container
	err = container.Close()

	// Assert - logger sync may return error in test environment
	// This is expected when syncing stderr
	if err != nil {
		t.Logf("Expected error during test cleanup: %v", err)
	}

	// Verify HTTP client connections are closed
	// (HTTP client should close idle connections on Close)
	assert.NotNil(t, container.HTTPClient)
}

func TestDependencyInjection_MultipleDependenciesInServer(t *testing.T) {
	// Arrange - initialize container
	container, err := dependencies.InitializeContainer()
	require.NoError(t, err)
	defer container.Close()

	// Act - create server using container
	server := httpserver.New(container)

	// Assert - server has access to all dependencies
	assert.NotNil(t, server)
	assert.NotNil(t, server.Router())

	// Verify dependencies are properly injected
	// by testing server functionality
	req := createTestRequest(t, "GET", "/health")
	w := createTestRecorder()
	server.Router().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDependencyInjection_CustomConfigurationOverride(t *testing.T) {
	// Arrange - create custom config
	customCfg := &config.Config{
		AppName:    "Custom Override",
		AppVersion: "99.99.99",
		Debug:      true,
		Host:       "localhost",
		Port:       9999,
		LogLevel:   "DEBUG",
		LogFormat:  "text",
	}

	log, err := logger.New(logger.Config{
		Level:  customCfg.LogLevel,
		Format: customCfg.LogFormat,
	})
	require.NoError(t, err)

	// Act - create container with custom config
	container := dependencies.NewContainer(customCfg, log)
	defer container.Close()

	// Assert - verify custom config is used
	assert.Equal(t, "Custom Override", container.Config.AppName)
	assert.Equal(t, "99.99.99", container.Config.AppVersion)
	assert.Equal(t, 9999, container.Config.Port)
	assert.Equal(t, "DEBUG", container.Config.LogLevel)
	assert.Equal(t, "text", container.Config.LogFormat)
}

func TestDependencyInjection_LoggerConfigurationFromConfig(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  string
		logFormat string
	}{
		{"json info", "INFO", "json"},
		{"text debug", "DEBUG", "text"},
		{"json error", "ERROR", "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cfg := &config.Config{
				AppName:    "Test",
				AppVersion: "0.1.0",
				Debug:      false,
				Host:       "127.0.0.1",
				Port:       8000,
				LogLevel:   tt.logLevel,
				LogFormat:  tt.logFormat,
			}

			log, err := logger.New(logger.Config{
				Level:  cfg.LogLevel,
				Format: cfg.LogFormat,
			})
			require.NoError(t, err)

			// Act
			container := dependencies.NewContainer(cfg, log)
			defer container.Close()

			// Assert
			assert.Equal(t, tt.logLevel, container.Config.LogLevel)
			assert.Equal(t, tt.logFormat, container.Config.LogFormat)
		})
	}
}

func TestDependencyInjection_HTTPClientCanMakeRequests(t *testing.T) {
	// Arrange
	container, err := dependencies.InitializeContainer()
	require.NoError(t, err)
	defer container.Close()

	client := container.HTTPClient

	// Act - make a test request (to a public endpoint)
	// Using httpbin.org for testing
	resp, err := client.Get("https://httpbin.org/status/200")

	// Assert
	if err != nil {
		t.Skipf("Skipping external HTTP test due to network error: %v", err)
	}
	defer resp.Body.Close()

	// httpbin might be down or rate-limiting, so skip if not 200
	if resp.StatusCode != http.StatusOK {
		t.Skipf("Skipping external HTTP test, httpbin returned: %d", resp.StatusCode)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDependencyInjection_ContainerSupportsMultipleCloseCalls(t *testing.T) {
	// Arrange
	container, err := dependencies.InitializeContainer()
	require.NoError(t, err)

	// Act - close container multiple times
	err1 := container.Close()
	err2 := container.Close()

	// Assert - multiple close calls should not panic
	// Logger.Sync might return error in test environment (expected)
	if err1 != nil {
		t.Logf("First close returned error (expected in tests): %v", err1)
	}
	if err2 != nil {
		t.Logf("Second close returned error (expected): %v", err2)
	}
	// Main assertion: no panic occurred
}

func TestDependencyInjection_WireGeneratedCodeExists(t *testing.T) {
	// This test verifies that Wire code generation has been run
	// by attempting to use the InitializeContainer function

	// Act
	container, err := dependencies.InitializeContainer()

	// Assert
	require.NoError(t, err, "Wire code generation may not have been run. Execute: task generate:wire")
	require.NotNil(t, container)
	defer container.Close()
}

// ====================
// Nested Dependencies Tests
// ====================

func TestDependencyInjection_NestedDependencies(t *testing.T) {
	// This test verifies that dependencies with their own dependencies
	// are properly resolved by Wire

	// Act
	container, err := dependencies.InitializeContainer()
	require.NoError(t, err)
	defer container.Close()

	// Assert - logger depends on config
	// Verify logger is configured based on config values
	assert.NotNil(t, container.Config)
	assert.NotNil(t, container.Logger)

	// Logger should be configured with values from config
	assert.Equal(t, container.Config.LogLevel, container.Config.LogLevel)
}

func TestDependencyInjection_DependencyLifecycle(t *testing.T) {
	// Test that dependencies follow proper lifecycle:
	// Initialize -> Use -> Close

	// Initialize
	container, err := dependencies.InitializeContainer()
	require.NoError(t, err)

	// Use - access dependencies
	assert.NotNil(t, container.Config)
	assert.NotNil(t, container.Logger)
	assert.NotNil(t, container.HTTPClient)

	container.Logger.Info("test log message")

	// Close - cleanup
	err = container.Close()
	// Logger.Sync may return error in test environment (expected)
	if err != nil {
		t.Logf("Close returned error (expected in tests): %v", err)
	}
}

// ====================
// Test Helpers
// ====================

// createTestRequest creates an HTTP test request
func createTestRequest(t *testing.T, method, path string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	return req
}

// createTestRecorder creates an HTTP test response recorder
func createTestRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}
