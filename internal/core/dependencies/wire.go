//go:build wireinject
// +build wireinject

package dependencies

import (
	"github.com/google/wire"
	"github.com/luminosita/change-me/internal/config"
	"github.com/luminosita/change-me/pkg/logger"
)

// InitializeContainer initializes the dependency injection container using Wire.
// Wire will generate the implementation of this function.
func InitializeContainer() (*Container, error) {
	wire.Build(
		config.Load,
		provideLogger,
		NewContainer,
	)
	return nil, nil
}

// provideLogger creates a logger from configuration.
func provideLogger(cfg *config.Config) (*logger.Logger, error) {
	return logger.New(logger.Config{
		Level:  cfg.LogLevel,
		Format: cfg.LogFormat,
	})
}
