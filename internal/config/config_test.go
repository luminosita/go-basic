package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_DefaultValues(t *testing.T) {
	// Clear environment variables
	clearEnvVars(t)

	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "CHANGE_ME", cfg.AppName)
	assert.Equal(t, "0.1.0", cfg.AppVersion)
	assert.False(t, cfg.Debug)
	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, 8000, cfg.Port)
	assert.Equal(t, "INFO", cfg.LogLevel)
	assert.Equal(t, "json", cfg.LogFormat)
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	// Set environment variables
	t.Setenv("APP_NAME", "Test Server")
	t.Setenv("APP_VERSION", "2.0.0")
	t.Setenv("DEBUG", "true")
	t.Setenv("PORT", "9000")
	t.Setenv("LOG_LEVEL", "DEBUG")
	t.Setenv("LOG_FORMAT", "text")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "Test Server", cfg.AppName)
	assert.Equal(t, "2.0.0", cfg.AppVersion)
	assert.True(t, cfg.Debug)
	assert.Equal(t, 9000, cfg.Port)
	assert.Equal(t, "DEBUG", cfg.LogLevel)
	assert.Equal(t, "text", cfg.LogFormat)
}

func TestLoad_ValidPort(t *testing.T) {
	tests := []struct {
		name string
		port string
		want int
	}{
		{"minimum port", "1", 1},
		{"common port", "8080", 8080},
		{"maximum port", "65535", 65535},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars(t)
			t.Setenv("PORT", tt.port)

			cfg, err := Load()
			require.NoError(t, err)
			assert.Equal(t, tt.want, cfg.Port)
		})
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port string
	}{
		{"port too low", "0"},
		{"port too high", "65536"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars(t)
			t.Setenv("PORT", tt.port)

			_, err := Load()
			assert.Error(t, err)
		})
	}
}

func TestLoad_LogLevelNormalization(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"lowercase debug", "debug", "DEBUG"},
		{"lowercase info", "info", "INFO"},
		{"uppercase warning", "WARNING", "WARNING"},
		{"mixed case", "Error", "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars(t)
			t.Setenv("LOG_LEVEL", tt.input)

			cfg, err := Load()
			require.NoError(t, err)
			assert.Equal(t, tt.want, cfg.LogLevel)
		})
	}
}

func TestLoad_ValidLogLevels(t *testing.T) {
	levels := []string{"DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			clearEnvVars(t)
			t.Setenv("LOG_LEVEL", level)

			cfg, err := Load()
			require.NoError(t, err)
			assert.Equal(t, level, cfg.LogLevel)
		})
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	clearEnvVars(t)
	t.Setenv("LOG_LEVEL", "INVALID")

	_, err := Load()
	assert.Error(t, err)
}

func TestLoad_ValidLogFormats(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{"json format", "json"},
		{"text format", "text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars(t)
			t.Setenv("LOG_FORMAT", tt.format)

			cfg, err := Load()
			require.NoError(t, err)
			assert.Equal(t, tt.format, cfg.LogFormat)
		})
	}
}

func TestLoad_InvalidLogFormat(t *testing.T) {
	clearEnvVars(t)
	t.Setenv("LOG_FORMAT", "invalid")

	_, err := Load()
	assert.Error(t, err)
}

func TestLoad_CustomAppName(t *testing.T) {
	clearEnvVars(t)
	t.Setenv("APP_NAME", "Custom FastAPI Server")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "Custom FastAPI Server", cfg.AppName)
}

func TestLoad_CustomHost(t *testing.T) {
	tests := []struct {
		name string
		host string
	}{
		{"localhost", "localhost"},
		{"127.0.0.1", "127.0.0.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars(t)
			t.Setenv("HOST", tt.host)

			cfg, err := Load()
			require.NoError(t, err)
			assert.Equal(t, tt.host, cfg.Host)
		})
	}
}

// clearEnvVars clears all config-related environment variables
func clearEnvVars(t *testing.T) {
	t.Helper()
	envVars := []string{
		"APP_NAME", "APP_VERSION", "DEBUG", "HOST", "PORT",
		"LOG_LEVEL", "LOG_FORMAT",
	}
	for _, key := range envVars {
		_ = os.Unsetenv(key)
	}
}
