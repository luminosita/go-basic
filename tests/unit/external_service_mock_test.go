package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/luminosita/change-me/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ====================
// External Service Example
// ====================

// ExternalAPIClient represents a client for an external API
type ExternalAPIClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewExternalAPIClient creates a new external API client
func NewExternalAPIClient(client *http.Client, baseURL string) *ExternalAPIClient {
	return &ExternalAPIClient{
		httpClient: client,
		baseURL:    baseURL,
	}
}

// GetUser fetches user data from external API
func (c *ExternalAPIClient) GetUser(ctx context.Context, userID string) (*UserResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/users/"+userID, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("user not found")
	}

	var user UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// CreateUser creates a new user via external API
func (c *ExternalAPIClient) CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/users", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New("failed to create user")
	}

	var user UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// UserResponse represents a user response from the API
type UserResponse struct {
	ID       string    `json:"id"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
	Created  time.Time `json:"created"`
}

// CreateUserRequest represents a user creation request
type CreateUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

// ====================
// Mock Tests for External Service
// ====================

func TestExternalAPIClient_GetUser_Success(t *testing.T) {
	// Arrange
	mockTransport := new(mocks.MockRoundTripper)

	expectedUser := UserResponse{
		ID:       "user123",
		Email:    "test@example.com",
		Username: "testuser",
		Created:  time.Now(),
	}

	responseBody, _ := json.Marshal(expectedUser)
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer(responseBody)),
		Header:     make(http.Header),
	}

	// Use AnythingOfType to match any request
	mockTransport.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)

	client := NewExternalAPIClient(mocks.NewMockHTTPClient(mockTransport), "https://api.example.com")

	// Act
	user, err := client.GetUser(context.Background(), "user123")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "user123", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "testuser", user.Username)

	mockTransport.AssertExpectations(t)
}

func TestExternalAPIClient_GetUser_NotFound(t *testing.T) {
	// Arrange
	mockTransport := new(mocks.MockRoundTripper)

	mockResponse := &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(bytes.NewBufferString(`{"error": "user not found"}`)),
		Header:     make(http.Header),
	}

	mockTransport.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)

	client := NewExternalAPIClient(mocks.NewMockHTTPClient(mockTransport), "https://api.example.com")

	// Act
	user, err := client.GetUser(context.Background(), "nonexistent")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "not found")

	mockTransport.AssertExpectations(t)
}

func TestExternalAPIClient_GetUser_NetworkError(t *testing.T) {
	// Arrange
	mockTransport := new(mocks.MockRoundTripper)

	networkErr := errors.New("network connection failed")
	mockTransport.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return((*http.Response)(nil), networkErr)

	client := NewExternalAPIClient(mocks.NewMockHTTPClient(mockTransport), "https://api.example.com")

	// Act
	user, err := client.GetUser(context.Background(), "user123")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "network")

	mockTransport.AssertExpectations(t)
}

func TestExternalAPIClient_CreateUser_Success(t *testing.T) {
	// Arrange
	mockTransport := new(mocks.MockRoundTripper)

	createReq := &CreateUserRequest{
		Email:    "newuser@example.com",
		Username: "newuser",
	}

	expectedUser := UserResponse{
		ID:       "user456",
		Email:    createReq.Email,
		Username: createReq.Username,
		Created:  time.Now(),
	}

	responseBody, _ := json.Marshal(expectedUser)
	mockResponse := &http.Response{
		StatusCode: http.StatusCreated,
		Body:       io.NopCloser(bytes.NewBuffer(responseBody)),
		Header:     make(http.Header),
	}

	mockTransport.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)

	client := NewExternalAPIClient(mocks.NewMockHTTPClient(mockTransport), "https://api.example.com")

	// Act
	user, err := client.CreateUser(context.Background(), createReq)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "user456", user.ID)
	assert.Equal(t, "newuser@example.com", user.Email)
	assert.Equal(t, "newuser", user.Username)

	mockTransport.AssertExpectations(t)
}

func TestExternalAPIClient_CreateUser_ValidationError(t *testing.T) {
	// Arrange
	mockTransport := new(mocks.MockRoundTripper)

	mockResponse := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(bytes.NewBufferString(`{"error": "invalid email"}`)),
		Header:     make(http.Header),
	}

	mockTransport.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)

	client := NewExternalAPIClient(mocks.NewMockHTTPClient(mockTransport), "https://api.example.com")

	createReq := &CreateUserRequest{
		Email:    "invalid-email",
		Username: "testuser",
	}

	// Act
	user, err := client.CreateUser(context.Background(), createReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to create")

	mockTransport.AssertExpectations(t)
}

func TestExternalAPIClient_CreateUser_Timeout(t *testing.T) {
	// Arrange
	mockTransport := new(mocks.MockRoundTripper)

	mockTransport.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(nil, &timeoutError{})

	client := NewExternalAPIClient(mocks.NewMockHTTPClient(mockTransport), "https://api.example.com")

	createReq := &CreateUserRequest{
		Email:    "test@example.com",
		Username: "testuser",
	}

	// Act
	user, err := client.CreateUser(context.Background(), createReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, isTimeout(err))

	mockTransport.AssertExpectations(t)
}

func TestExternalAPIClient_CreateUser_InvalidJSON(t *testing.T) {
	// Arrange
	mockTransport := new(mocks.MockRoundTripper)

	mockResponse := &http.Response{
		StatusCode: http.StatusCreated,
		Body:       io.NopCloser(bytes.NewBufferString(`invalid json`)),
		Header:     make(http.Header),
	}

	mockTransport.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(mockResponse, nil)

	client := NewExternalAPIClient(mocks.NewMockHTTPClient(mockTransport), "https://api.example.com")

	createReq := &CreateUserRequest{
		Email:    "test@example.com",
		Username: "testuser",
	}

	// Act
	user, err := client.CreateUser(context.Background(), createReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)

	mockTransport.AssertExpectations(t)
}

// ====================
// Test Context Cancellation
// ====================

func TestExternalAPIClient_GetUser_ContextCanceled(t *testing.T) {
	// Arrange
	mockTransport := new(mocks.MockRoundTripper)

	// Setup mock to return context canceled error
	mockTransport.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return((*http.Response)(nil), context.Canceled)

	// Create canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := NewExternalAPIClient(mocks.NewMockHTTPClient(mockTransport), "https://api.example.com")

	// Act
	user, err := client.GetUser(ctx, "user123")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)

	// Context cancellation or network error is expected
	// Note: Mock may or may not be called depending on when cancellation happens
}
