package unit

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/luminosita/change-me/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ====================
// Mock HTTP Client Tests
// ====================

func TestMockHTTPClient_GetRequest(t *testing.T) {
	// Arrange - create mock transport
	mockTransport := new(mocks.MockRoundTripper)

	// Setup expected response
	expectedResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"status": "ok"}`)),
		Header:     make(http.Header),
	}
	expectedResponse.Header.Set("Content-Type", "application/json")

	// Configure mock to return expected response for any request
	mockTransport.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(expectedResponse, nil)

	// Create HTTP client with mock transport
	client := mocks.NewMockHTTPClient(mockTransport)

	// Act - make GET request
	req, err := http.NewRequest("GET", "https://api.example.com/data", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	// Verify body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"status": "ok"}`, string(body))

	// Verify mock was called
	mockTransport.AssertExpectations(t)
}

func TestMockHTTPClient_RequestWithError(t *testing.T) {
	// Arrange - create mock transport
	mockTransport := new(mocks.MockRoundTripper)

	// Configure mock to return error for any request
	mockTransport.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(nil, assert.AnError)

	client := mocks.NewMockHTTPClient(mockTransport)

	// Act - make request that should fail
	req, err := http.NewRequest("GET", "https://api.example.com/error", nil)
	require.NoError(t, err)

	_, err = client.Do(req)

	// Assert - should receive error
	assert.Error(t, err)
	mockTransport.AssertExpectations(t)
}

func TestMockHTTPClient_PostRequest(t *testing.T) {
	// Arrange
	mockTransport := new(mocks.MockRoundTripper)

	expectedResponse := &http.Response{
		StatusCode: http.StatusCreated,
		Body:       io.NopCloser(bytes.NewBufferString(`{"id": "123", "created": true}`)),
		Header:     make(http.Header),
	}

	mockTransport.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(expectedResponse, nil)

	client := mocks.NewMockHTTPClient(mockTransport)

	// Act - make POST request
	req, err := http.NewRequest("POST", "https://api.example.com/items", bytes.NewBufferString(`{"name": "test"}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"created": true`)

	mockTransport.AssertExpectations(t)
}

func TestMockHTTPClient_Timeout(t *testing.T) {
	// Arrange
	mockTransport := new(mocks.MockRoundTripper)

	// Simulate timeout error for any request
	mockTransport.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(nil, &timeoutError{})

	client := mocks.NewMockHTTPClient(mockTransport)

	// Act
	req, err := http.NewRequest("GET", "https://api.example.com/slow", nil)
	require.NoError(t, err)

	_, err = client.Do(req)

	// Assert
	assert.Error(t, err)
	assert.True(t, isTimeout(err), "Expected timeout error")

	mockTransport.AssertExpectations(t)
}

func TestMockHTTPClient_MultipleRequests(t *testing.T) {
	// Arrange
	mockTransport := new(mocks.MockRoundTripper)

	// Setup multiple responses
	response1 := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"page": 1}`)),
	}
	response2 := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"page": 2}`)),
	}

	// Setup mock to return different responses on consecutive calls
	mockTransport.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(response1, nil).Once()
	mockTransport.On("RoundTrip", mock.AnythingOfType("*http.Request")).Return(response2, nil).Once()

	client := mocks.NewMockHTTPClient(mockTransport)

	// Act - make two different requests
	req1, _ := http.NewRequest("GET", "https://api.example.com/page/1", nil)
	resp1, err1 := client.Do(req1)
	require.NoError(t, err1)

	req2, _ := http.NewRequest("GET", "https://api.example.com/page/2", nil)
	resp2, err2 := client.Do(req2)
	require.NoError(t, err2)

	// Assert
	assert.Equal(t, http.StatusOK, resp1.StatusCode)
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	body1, _ := io.ReadAll(resp1.Body)
	body2, _ := io.ReadAll(resp2.Body)

	assert.Contains(t, string(body1), `"page": 1`)
	assert.Contains(t, string(body2), `"page": 2`)

	mockTransport.AssertExpectations(t)
}

// ====================
// Test Helpers
// ====================

// timeoutError simulates a timeout error
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

// isTimeout checks if an error is a timeout error
func isTimeout(err error) bool {
	type timeout interface {
		Timeout() bool
	}
	if te, ok := err.(timeout); ok {
		return te.Timeout()
	}
	return false
}
