package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCommitMessage(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Decode request body
		var reqBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NoError(t, err)

		// Verify request body
		assert.Equal(t, float64(500), reqBody["max_tokens"])
		assert.Equal(t, float64(0.7), reqBody["temperature"])
		assert.Contains(t, reqBody["prompt"], "Please analyze the following code changes")

		// Send response
		response := map[string]interface{}{
			"response": "feat: add new feature\n\nThis is a detailed description of the changes.",
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		assert.NoError(t, err)
	}))
	defer server.Close()

	// Override llama API URL for testing
	originalURL := llamaAPIURL
	err := SetLlamaAPIURL(server.URL)
	assert.NoError(t, err)
	defer func() {
		err := SetLlamaAPIURL(originalURL)
		assert.NoError(t, err)
	}()

	// Test cases
	tests := []struct {
		name     string
		diffs    []string
		expected string
	}{
		{
			name:     "empty diffs",
			diffs:    []string{},
			expected: "feat: add new feature\n\nThis is a detailed description of the changes.",
		},
		{
			name:     "with diffs",
			diffs:    []string{"diff content"},
			expected: "feat: add new feature\n\nThis is a detailed description of the changes.",
		},
		{
			name:     "multiple diffs",
			diffs:    []string{"diff1", "diff2"},
			expected: "feat: add new feature\n\nThis is a detailed description of the changes.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, err := GenerateCommitMessage(tt.diffs)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, message)
		})
	}
}

func TestGenerateCommitMessageError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Override llama API URL for testing
	originalURL := llamaAPIURL
	err := SetLlamaAPIURL(server.URL)
	assert.NoError(t, err)
	defer func() {
		err := SetLlamaAPIURL(originalURL)
		assert.NoError(t, err)
	}()

	// Test cases
	tests := []struct {
		name          string
		serverHandler http.HandlerFunc
		expectedError string
	}{
		{
			name: "server error",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedError: "llama API returned status code 500",
		},
		{
			name: "invalid JSON response",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte("invalid json"))
				assert.NoError(t, err)
			},
			expectedError: "failed to decode llama API response",
		},
		{
			name: "missing response field",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				response := map[string]interface{}{
					"other_field": "some value",
				}
				err := json.NewEncoder(w).Encode(response)
				assert.NoError(t, err)
			},
			expectedError: "missing response field in llama API response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server with the specific handler
			server := httptest.NewServer(tt.serverHandler)
			defer server.Close()

			// Override llama API URL for testing
			originalURL := llamaAPIURL
			err := SetLlamaAPIURL(server.URL)
			assert.NoError(t, err)
			defer func() {
				err := SetLlamaAPIURL(originalURL)
				assert.NoError(t, err)
			}()

			_, err = GenerateCommitMessage([]string{"diff"})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestGenerateCommitMessageDefaultType(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"response": "This is a commit message without a type",
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		assert.NoError(t, err)
	}))
	defer server.Close()

	// Override llama API URL for testing
	originalURL := llamaAPIURL
	err := SetLlamaAPIURL(server.URL)
	assert.NoError(t, err)
	defer func() {
		err := SetLlamaAPIURL(originalURL)
		assert.NoError(t, err)
	}()

	message, err := GenerateCommitMessage([]string{"diff"})
	assert.NoError(t, err)
	assert.Equal(t, "feat: This is a commit message without a type", message)
}
