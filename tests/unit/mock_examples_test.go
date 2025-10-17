package unit

import (
	"testing"

	"github.com/luminosita/change-me/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ====================
// Sample Data Fixture Examples
// ====================

func TestSampleUserData_DefaultValues(t *testing.T) {
	// Arrange & Act - use default fixture
	user := mocks.NewSampleUser()

	// Assert - verify default values
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "Test User", user.FullName)
	assert.True(t, user.IsActive)
}

func TestSampleUserData_WithCustomValues(t *testing.T) {
	// Arrange & Act - use fixture with custom values
	user := mocks.NewSampleUser(
		mocks.WithUserID(42),
		mocks.WithUserEmail("custom@example.com"),
		mocks.WithUsername("customuser"),
	)

	// Assert - verify custom values
	assert.Equal(t, 42, user.ID)
	assert.Equal(t, "custom@example.com", user.Email)
	assert.Equal(t, "customuser", user.Username)
	assert.Equal(t, "Test User", user.FullName) // Default
	assert.True(t, user.IsActive)                // Default
}

func TestSampleUserData_MultipleUsers(t *testing.T) {
	// Arrange & Act - create multiple users with factory
	user1 := mocks.NewSampleUser(
		mocks.WithUserID(1),
		mocks.WithUserEmail("user1@example.com"),
	)
	user2 := mocks.NewSampleUser(
		mocks.WithUserID(2),
		mocks.WithUserEmail("user2@example.com"),
	)
	user3 := mocks.NewSampleUser(
		mocks.WithUserID(3),
		mocks.WithUserEmail("user3@example.com"),
	)

	// Assert - each user has unique data
	assert.NotEqual(t, user1.ID, user2.ID)
	assert.NotEqual(t, user2.ID, user3.ID)
	assert.NotEqual(t, user1.Email, user2.Email)
	assert.NotEqual(t, user2.Email, user3.Email)
}

func TestSampleUserData_InactiveUser(t *testing.T) {
	// Arrange & Act - create inactive user
	user := mocks.NewSampleUser(
		mocks.WithUserActive(false),
	)

	// Assert
	assert.False(t, user.IsActive)
}

// ====================
// Config Factory Examples
// ====================

func TestConfigFactory_DefaultValues(t *testing.T) {
	// Arrange & Act
	cfg := mocks.NewTestConfig()

	// Assert
	assert.Equal(t, "TestApp", cfg.AppName)
	assert.Equal(t, "0.1.0", cfg.AppVersion)
	assert.True(t, cfg.Debug)
	assert.Equal(t, "127.0.0.1", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "INFO", cfg.LogLevel)
	assert.Equal(t, "json", cfg.LogFormat)
}

func TestConfigFactory_WithCustomValues(t *testing.T) {
	// Arrange & Act
	cfg := mocks.NewTestConfig(
		mocks.WithAppName("MyCustomApp"),
		mocks.WithPort(9000),
		mocks.WithLogLevel("DEBUG"),
		mocks.WithDebug(false),
	)

	// Assert
	assert.Equal(t, "MyCustomApp", cfg.AppName)
	assert.Equal(t, 9000, cfg.Port)
	assert.Equal(t, "DEBUG", cfg.LogLevel)
	assert.False(t, cfg.Debug)
	assert.Equal(t, "0.1.0", cfg.AppVersion) // Default
}

func TestConfigFactory_ProductionConfig(t *testing.T) {
	// Arrange & Act - create production-like config
	cfg := mocks.NewTestConfig(
		mocks.WithAppName("ProductionApp"),
		mocks.WithDebug(false),
		mocks.WithLogLevel("ERROR"),
		mocks.WithPort(443),
	)

	// Assert
	assert.Equal(t, "ProductionApp", cfg.AppName)
	assert.False(t, cfg.Debug)
	assert.Equal(t, "ERROR", cfg.LogLevel)
	assert.Equal(t, 443, cfg.Port)
}

// ====================
// Mock Logger Examples
// ====================

func TestMockLogger_InfoCalled(t *testing.T) {
	// Arrange
	mockLogger := mocks.NewMockLogger()
	mockLogger.On("Info", "test message").Return()

	// Act
	mockLogger.Info("test message")

	// Assert
	mockLogger.AssertExpectations(t)
	mockLogger.AssertCalled(t, "Info", "test message")
}

func TestMockLogger_InfowWithFields(t *testing.T) {
	// Arrange
	mockLogger := mocks.NewMockLogger()
	mockLogger.On("Infow", "user created", "user_id", 123, "email", "test@example.com").Return()

	// Act
	mockLogger.Infow("user created", "user_id", 123, "email", "test@example.com")

	// Assert
	mockLogger.AssertExpectations(t)
	mockLogger.AssertCalled(t, "Infow", "user created", "user_id", 123, "email", "test@example.com")
}

func TestMockLogger_ErrorCalled(t *testing.T) {
	// Arrange
	mockLogger := mocks.NewMockLogger()
	mockLogger.On("Error", "operation failed", "error", "database connection timeout").Return()

	// Act
	mockLogger.Error("operation failed", "error", "database connection timeout")

	// Assert
	mockLogger.AssertExpectations(t)
	mockLogger.AssertCalled(t, "Error", "operation failed", "error", "database connection timeout")
}

func TestMockLogger_SyncReturnsError(t *testing.T) {
	// Arrange
	mockLogger := mocks.NewMockLogger()
	mockLogger.On("Sync").Return(assert.AnError)

	// Act
	err := mockLogger.Sync()

	// Assert
	assert.Error(t, err)
	mockLogger.AssertExpectations(t)
}

func TestMockLogger_SyncSuccess(t *testing.T) {
	// Arrange
	mockLogger := mocks.NewMockLogger()
	mockLogger.On("Sync").Return(nil)

	// Act
	err := mockLogger.Sync()

	// Assert
	assert.NoError(t, err)
	mockLogger.AssertExpectations(t)
}

// ====================
// Real Logger Factory Examples
// ====================

func TestLoggerFactory_CreatesRealLogger(t *testing.T) {
	// Arrange & Act
	logger, err := mocks.NewTestLogger()

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, logger)

	// Verify logger is functional
	assert.NotPanics(t, func() {
		logger.Info("test log message")
	})
}

// ====================
// Parametrized Test Examples
// ====================

func TestSampleUserData_DifferentActiveStatus(t *testing.T) {
	tests := []struct {
		name     string
		userID   int
		isActive bool
	}{
		{"active user 1", 1, true},
		{"inactive user 2", 2, false},
		{"active user 3", 3, true},
		{"inactive user 4", 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			user := mocks.NewSampleUser(
				mocks.WithUserID(tt.userID),
				mocks.WithUserActive(tt.isActive),
			)

			// Assert
			assert.Equal(t, tt.userID, user.ID)
			assert.Equal(t, tt.isActive, user.IsActive)
		})
	}
}

func TestConfigFactory_DifferentLogLevels(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
	}{
		{"debug level", "DEBUG"},
		{"info level", "INFO"},
		{"warning level", "WARNING"},
		{"error level", "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			cfg := mocks.NewTestConfig(
				mocks.WithLogLevel(tt.logLevel),
			)

			// Assert
			assert.Equal(t, tt.logLevel, cfg.LogLevel)
		})
	}
}

func TestSampleUserData_EmailValidation(t *testing.T) {
	tests := []struct {
		name         string
		email        string
		expectValid  bool
		validateFunc func(string) bool
	}{
		{
			name:        "valid email",
			email:       "user@example.com",
			expectValid: true,
			validateFunc: func(e string) bool {
				return len(e) > 0 && contains(e, "@") && contains(e, ".")
			},
		},
		{
			name:        "another valid email",
			email:       "test.user@company.co.uk",
			expectValid: true,
			validateFunc: func(e string) bool {
				return len(e) > 0 && contains(e, "@") && contains(e, ".")
			},
		},
		{
			name:        "invalid email - no @",
			email:       "invalid.email.com",
			expectValid: false,
			validateFunc: func(e string) bool {
				return len(e) > 0 && contains(e, "@") && contains(e, ".")
			},
		},
		{
			name:        "invalid email - no domain",
			email:       "user@",
			expectValid: false,
			validateFunc: func(e string) bool {
				parts := splitString(e, "@")
				return len(parts) == 2 && contains(parts[1], ".")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			user := mocks.NewSampleUser(
				mocks.WithUserEmail(tt.email),
			)

			// Assert
			isValid := tt.validateFunc(user.Email)
			assert.Equal(t, tt.expectValid, isValid,
				"Email %s validation result mismatch", tt.email)
		})
	}
}

// ====================
// Test Helpers
// ====================

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i = start - 1
		}
	}
	result = append(result, s[start:])
	return result
}
