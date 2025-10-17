package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luminosita/change-me/internal/core/dependencies"
	"github.com/luminosita/change-me/internal/interfaces/http/handlers"
	"github.com/luminosita/change-me/internal/interfaces/http/middleware"
)

// Server represents the HTTP server.
type Server struct {
	router    *gin.Engine
	container *dependencies.Container
}

// New creates a new HTTP server with all routes and middleware configured.
func New(container *dependencies.Container) *Server {
	// Set Gin mode based on debug setting
	if !container.Config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Register middleware
	router.Use(gin.Recovery()) // Panic recovery
	router.Use(middleware.CORS())
	router.Use(middleware.Logger(container.Logger))

	// Health check handler
	healthHandler := handlers.NewHealthHandler(container.Config.AppVersion)
	router.GET("/health", healthHandler.Check)

	return &Server{
		router:    router,
		container: container,
	}
}

// Router returns the underlying Gin router for testing.
func (s *Server) Router() *gin.Engine {
	return s.router
}

// Start starts the HTTP server with graceful shutdown support.
func (s *Server) Start() error {
	cfg := s.container.Config
	log := s.container.Logger

	// Server address
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// Create HTTP server
	srv := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Log startup information
	log.Infow("application_startup",
		"app_name", cfg.AppName,
		"version", cfg.AppVersion,
		"host", cfg.Host,
		"port", cfg.Port,
		"debug", cfg.Debug,
		"log_level", cfg.LogLevel,
		"log_format", cfg.LogFormat,
	)

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalw("server_failed", "error", err)
		}
	}()

	log.Infow("application_startup_complete", "address", addr)

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Infow("application_shutdown_started")

	// Shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Errorw("server_shutdown_error", "error", err)
		return err
	}

	// Close dependencies
	if err := s.container.Close(); err != nil {
		log.Errorw("dependencies_close_error", "error", err)
		return err
	}

	log.Infow("application_shutdown_complete")
	return nil
}
