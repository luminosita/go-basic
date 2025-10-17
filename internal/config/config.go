package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
// Configuration values are loaded from:
// 1. Environment variables (highest priority)
// 2. .env file (development default)
// 3. Default values (fallback)
type Config struct {
	// Application metadata
	AppName    string `mapstructure:"APP_NAME" validate:"required"`
	AppVersion string `mapstructure:"APP_VERSION" validate:"required"`
	Debug      bool   `mapstructure:"DEBUG"`

	// Server configuration
	Host string `mapstructure:"HOST" validate:"required"`
	Port int    `mapstructure:"PORT" validate:"required,min=1,max=65535"`

	// Logging configuration
	LogLevel  string `mapstructure:"LOG_LEVEL" validate:"required,oneof=DEBUG INFO WARNING ERROR CRITICAL"`
	LogFormat string `mapstructure:"LOG_FORMAT" validate:"required,oneof=json text"`
}

// Load reads configuration from environment variables and .env file.
// It returns a validated Config instance or an error if validation fails.
//
// Configuration precedence:
// 1. Environment variables (highest)
// 2. .env file
// 3. Default values (lowest)
func Load() (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("APP_NAME", "CHANGE_ME")
	v.SetDefault("APP_VERSION", "0.1.0")
	v.SetDefault("DEBUG", false)
	v.SetDefault("HOST", "0.0.0.0")
	v.SetDefault("PORT", 8000)
	v.SetDefault("LOG_LEVEL", "INFO")
	v.SetDefault("LOG_FORMAT", "json")

	// Read from .env file (optional, won't error if missing)
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AddConfigPath("./")

	// Read config file (ignore error if file doesn't exist)
	_ = v.ReadInConfig()

	// Environment variables override file config
	v.AutomaticEnv()

	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Normalize log level to uppercase
	cfg.LogLevel = strings.ToUpper(cfg.LogLevel)

	// Validate configuration
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}
