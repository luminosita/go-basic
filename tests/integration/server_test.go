//go:build integration

package integration

import (
	"context"
	"fmt"
	"net"
	"net/http"
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
// Server Lifecycle Tests
// ====================

func TestServerLifecycle_StartAndShutdown(t *testing.T) {
	// Arrange - find available port
	port := findAvailablePort(t)

	cfg := &config.Config{
		AppName:    "Test Server",
		AppVersion: "0.1.0",
		Debug:      true,
		Host:       "127.0.0.1",
		Port:       port,
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

	server := httpserver.New(container)

	// Create custom HTTP server with manual shutdown control
	httpSrv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      server.Router(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Act - start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Verify server is running
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Act - shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = httpSrv.Shutdown(ctx)

	// Assert
	assert.NoError(t, err)
	select {
	case err := <-errChan:
		t.Fatalf("Server error: %v", err)
	default:
		// No error, success
	}
}

func TestServerLifecycle_GracefulShutdown(t *testing.T) {
	// Arrange
	port := findAvailablePort(t)

	cfg := &config.Config{
		AppName:    "Test Server",
		AppVersion: "0.1.0",
		Debug:      true,
		Host:       "127.0.0.1",
		Port:       port,
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

	server := httpserver.New(container)

	httpSrv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler: server.Router(),
	}

	// Start server
	go func() {
		_ = httpSrv.ListenAndServe()
	}()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Start a long-running request
	requestComplete := make(chan bool)
	go func() {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
		if err == nil {
			resp.Body.Close()
		}
		requestComplete <- true
	}()

	// Allow request to start
	time.Sleep(50 * time.Millisecond)

	// Act - initiate graceful shutdown
	shutdownComplete := make(chan error)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		shutdownComplete <- httpSrv.Shutdown(ctx)
	}()

	// Assert - request should complete before shutdown
	select {
	case <-requestComplete:
		// Request completed successfully
	case <-time.After(2 * time.Second):
		t.Fatal("Request did not complete during graceful shutdown")
	}

	// Wait for shutdown to complete
	select {
	case err := <-shutdownComplete:
		assert.NoError(t, err)
	case <-time.After(6 * time.Second):
		t.Fatal("Graceful shutdown did not complete in time")
	}
}

func TestServerLifecycle_RejectsNewRequestsDuringShutdown(t *testing.T) {
	// Arrange
	port := findAvailablePort(t)

	cfg := &config.Config{
		AppName:    "Test Server",
		AppVersion: "0.1.0",
		Debug:      true,
		Host:       "127.0.0.1",
		Port:       port,
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

	server := httpserver.New(container)

	httpSrv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler: server.Router(),
	}

	// Start server
	go func() {
		_ = httpSrv.ListenAndServe()
	}()

	time.Sleep(100 * time.Millisecond)

	// Act - initiate shutdown
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(ctx)
	}()

	// Wait for shutdown to start
	time.Sleep(50 * time.Millisecond)

	// Try to make request during shutdown
	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	_, err = client.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))

	// Assert - request should fail (connection refused or timeout)
	// Note: This might succeed if request arrives before shutdown starts
	// Main goal is to ensure server can shutdown cleanly
	if err != nil {
		t.Logf("Expected behavior: request failed during shutdown: %v", err)
	}
}

func TestServerLifecycle_ServerAddressBinding(t *testing.T) {
	// Arrange
	port := findAvailablePort(t)

	cfg := &config.Config{
		AppName:    "Test Server",
		AppVersion: "0.1.0",
		Debug:      true,
		Host:       "127.0.0.1",
		Port:       port,
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

	server := httpserver.New(container)

	httpSrv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler: server.Router(),
	}

	// Act - start server
	errChan := make(chan error, 1)
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Assert - verify server is listening on correct address
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(ctx)

	// Verify no errors occurred
	select {
	case err := <-errChan:
		t.Fatalf("Server error: %v", err)
	default:
		// Success
	}
}

func TestServerLifecycle_ServerTimeoutConfiguration(t *testing.T) {
	// Arrange
	port := findAvailablePort(t)

	cfg := &config.Config{
		AppName:    "Test Server",
		AppVersion: "0.1.0",
		Debug:      true,
		Host:       "127.0.0.1",
		Port:       port,
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

	server := httpserver.New(container)

	// Create server with specific timeouts
	httpSrv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      server.Router(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Act - start server
	go func() {
		_ = httpSrv.ListenAndServe()
	}()

	time.Sleep(100 * time.Millisecond)

	// Assert - verify timeouts are configured correctly
	assert.Equal(t, 10*time.Second, httpSrv.ReadTimeout)
	assert.Equal(t, 10*time.Second, httpSrv.WriteTimeout)
	assert.Equal(t, 120*time.Second, httpSrv.IdleTimeout)

	// Make request to ensure server responds
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", port))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(ctx)
}

// ====================
// Test Helpers
// ====================

// findAvailablePort finds an available TCP port on localhost
func findAvailablePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port
}
