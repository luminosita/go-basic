package main

import (
	"log"

	"github.com/luminosita/change-me/internal/core/dependencies"
	httpserver "github.com/luminosita/change-me/internal/interfaces/http"
)

// @title CHANGE_ME API
// @version 0.1.0
// @description Go HTTP server with health check, logging, and dependency injection
// @host localhost:8000
// @BasePath /
func main() {
	// Initialize dependency container with Wire
	container, err := dependencies.InitializeContainer()
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}

	// Create and start HTTP server
	server := httpserver.New(container)
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
