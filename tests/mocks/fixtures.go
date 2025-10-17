// Package mocks provides mock implementations and test fixtures for unit testing.
//
// This package contains reusable mock objects, test data factories, and helper
// functions to simplify writing unit tests with mocked dependencies.
package mocks

import (
	"net/http"
	"time"

	"github.com/luminosita/change-me/internal/config"
	"github.com/luminosita/change-me/pkg/logger"
	"github.com/stretchr/testify/mock"
)

// ====================
// Mock HTTP Transport
// ====================

// MockRoundTripper is a mock implementation of http.RoundTripper for testing HTTP clients.
type MockRoundTripper struct {
	mock.Mock
}

// RoundTrip implements the http.RoundTripper interface.
func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

// ====================
// Mock Logger
// ====================

// MockLogger is a mock implementation of logger methods for testing.
type MockLogger struct {
	mock.Mock
}

// Info logs an info message.
func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	args := []interface{}{msg}
	args = append(args, keysAndValues...)
	m.Called(args...)
}

// Infow logs an info message with structured fields.
func (m *MockLogger) Infow(msg string, keysAndValues ...interface{}) {
	args := []interface{}{msg}
	args = append(args, keysAndValues...)
	m.Called(args...)
}

// Error logs an error message.
func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	args := []interface{}{msg}
	args = append(args, keysAndValues...)
	m.Called(args...)
}

// Errorw logs an error message with structured fields.
func (m *MockLogger) Errorw(msg string, keysAndValues ...interface{}) {
	args := []interface{}{msg}
	args = append(args, keysAndValues...)
	m.Called(args...)
}

// Sync flushes any buffered log entries.
func (m *MockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}

// ====================
// Test Data Factories
// ====================

// NewTestConfig creates a test configuration with default values.
// Use functional options to override specific fields.
func NewTestConfig(opts ...func(*config.Config)) *config.Config {
	cfg := &config.Config{
		AppName:    "TestApp",
		AppVersion: "0.1.0",
		Debug:      true,
		Host:       "127.0.0.1",
		Port:       8080,
		LogLevel:   "INFO",
		LogFormat:  "json",
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// WithAppName sets the app name in config.
func WithAppName(name string) func(*config.Config) {
	return func(c *config.Config) {
		c.AppName = name
	}
}

// WithPort sets the port in config.
func WithPort(port int) func(*config.Config) {
	return func(c *config.Config) {
		c.Port = port
	}
}

// WithLogLevel sets the log level in config.
func WithLogLevel(level string) func(*config.Config) {
	return func(c *config.Config) {
		c.LogLevel = level
	}
}

// WithDebug sets the debug flag in config.
func WithDebug(debug bool) func(*config.Config) {
	return func(c *config.Config) {
		c.Debug = debug
	}
}

// NewTestLogger creates a real logger instance for testing.
func NewTestLogger() (*logger.Logger, error) {
	return logger.New(logger.Config{
		Level:  "INFO",
		Format: "json",
	})
}

// NewMockLogger creates a mock logger for testing.
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

// ====================
// Mock HTTP Client
// ====================

// NewMockHTTPClient creates an HTTP client with a mock transport.
func NewMockHTTPClient(transport *MockRoundTripper) *http.Client {
	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}

// ====================
// Sample Data Fixtures
// ====================

// SampleUserData represents sample user data for testing.
type SampleUserData struct {
	ID       int
	Email    string
	Username string
	FullName string
	IsActive bool
}

// NewSampleUser creates sample user data with defaults.
func NewSampleUser(opts ...func(*SampleUserData)) *SampleUserData {
	user := &SampleUserData{
		ID:       1,
		Email:    "test@example.com",
		Username: "testuser",
		FullName: "Test User",
		IsActive: true,
	}

	for _, opt := range opts {
		opt(user)
	}

	return user
}

// WithUserID sets the user ID.
func WithUserID(id int) func(*SampleUserData) {
	return func(u *SampleUserData) {
		u.ID = id
	}
}

// WithUserEmail sets the user email.
func WithUserEmail(email string) func(*SampleUserData) {
	return func(u *SampleUserData) {
		u.Email = email
	}
}

// WithUsername sets the username.
func WithUsername(username string) func(*SampleUserData) {
	return func(u *SampleUserData) {
		u.Username = username
	}
}

// WithUserActive sets the active status.
func WithUserActive(active bool) func(*SampleUserData) {
	return func(u *SampleUserData) {
		u.IsActive = active
	}
}
