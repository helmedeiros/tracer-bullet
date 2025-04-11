package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
		assert.Contains(t, reqBody["prompt"], "You are an expert at writing clear and descriptive commit messages")

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

func TestPromptFormat(t *testing.T) {
	tests := []struct {
		name           string
		diffs          []string
		expectedFormat string
	}{
		{
			name: "feature addition",
			diffs: []string{
				"File: internal/api/user.go\n" +
					"@@ -0,0 +1,10 @@\n" +
					"+package api\n" +
					"+\n" +
					"+type User struct {\n" +
					"+    ID       string\n" +
					"+    Username string\n" +
					"+    Email    string\n" +
					"+}\n" +
					"+\n" +
					"+func NewUser(id, username, email string) *User {\n" +
					"+    return &User{\n" +
					"+        ID:       id,\n" +
					"+        Username: username,\n" +
					"+        Email:    email,\n" +
					"+    }\n" +
					"+}",
			},
			expectedFormat: "feat(api): add user model and constructor\n\nAdd User struct with basic fields and constructor function.\n- Define User struct with ID, Username, and Email fields\n- Implement NewUser constructor function\n- Add package documentation",
		},
		{
			name: "bug fix",
			diffs: []string{
				"File: internal/utils/validation.go\n" +
					"@@ -10,6 +10,7 @@ func ValidateEmail(email string) bool {\n" +
					"     if len(email) == 0 {\n" +
					"         return false\n" +
					"     }\n" +
					"+    // Add check for @ symbol\n" +
					"+    if !strings.Contains(email, \"@\") {\n" +
					"+        return false\n" +
					"+    }\n" +
					"     return true\n" +
					" }",
			},
			expectedFormat: "fix(utils): add email validation check\n\nFix email validation to properly check for @ symbol.\n- Add check for @ symbol in email address\n- Improve validation logic\n- Prevent invalid email formats",
		},
		{
			name: "refactoring",
			diffs: []string{
				"File: internal/core/service.go\n" +
					"@@ -15,20 +15,25 @@ type Service struct {\n" +
					"     db *DB\n" +
					" }\n" +
					" \n" +
					"-func (s *Service) Process(data string) error {\n" +
					"+func (s *Service) Process(data string) (Result, error) {\n" +
					"     // Validate input\n" +
					"     if err := validate(data); err != nil {\n" +
					"         return err\n" +
					"+        return Result{}, err\n" +
					"     }\n" +
					" \n" +
					"     // Process data\n" +
					"-    result := process(data)\n" +
					"+    result, err := process(data)\n" +
					"     if err != nil {\n" +
					"         return err\n" +
					"+        return Result{}, err\n" +
					"     }\n" +
					" \n" +
					"-    return nil\n" +
					"+    return result, nil\n" +
					" }",
			},
			expectedFormat: "refactor(core): improve error handling in service\n\nRefactor service to use proper error handling and return values.\n- Change Process method to return Result and error\n- Update error handling to return Result struct\n- Improve method signature for better error handling",
		},
		{
			name: "documentation",
			diffs: []string{
				"File: README.md\n" +
					"@@ -1,3 +1,10 @@\n" +
					" # Project Name\n" +
					" \n" +
					"+## Installation\n" +
					"+\n" +
					"+```bash\n" +
					"+go get github.com/example/project\n" +
					"+```\n" +
					"+\n" +
					"+## Usage\n" +
					"+\n" +
					"+```go\n" +
					"+import \"github.com/example/project\"\n" +
					"+```",
			},
			expectedFormat: "docs: add installation and usage instructions\n\nAdd documentation for project installation and basic usage.\n- Add installation instructions with go get command\n- Include basic usage example with import statement\n- Improve README structure",
		},
		{
			name: "test addition",
			diffs: []string{
				"File: internal/api/user_test.go\n" +
					"@@ -0,0 +1,20 @@\n" +
					"+package api\n" +
					"+\n" +
					"+import (\n" +
					"+    \"testing\"\n" +
					"+    \"github.com/stretchr/testify/assert\"\n" +
					"+)\n" +
					"+\n" +
					"+func TestNewUser(t *testing.T) {\n" +
					"+    user := NewUser(\"123\", \"testuser\", \"test@example.com\")\n" +
					"+    assert.Equal(t, \"123\", user.ID)\n" +
					"+    assert.Equal(t, \"testuser\", user.Username)\n" +
					"+    assert.Equal(t, \"test@example.com\", user.Email)\n" +
					"+}",
			},
			expectedFormat: "test(api): add user constructor tests\n\nAdd test coverage for NewUser constructor function.\n- Create test file for user package\n- Implement TestNewUser function\n- Add assertions for all user fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				assert.Equal(t, "llama3", reqBody["model"])
				assert.Equal(t, float64(500), reqBody["max_tokens"])
				assert.Equal(t, float64(0.7), reqBody["temperature"])

				// Verify prompt contains our examples and format
				prompt := reqBody["prompt"].(string)
				assert.Contains(t, prompt, "You are an expert at writing clear and descriptive commit messages")
				assert.Contains(t, prompt, "The commit message MUST follow this exact format")
				assert.Contains(t, prompt, "Examples of good commit messages")
				assert.Contains(t, prompt, tt.diffs[0])

				// Send response
				response := map[string]interface{}{
					"response": tt.expectedFormat,
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

			// Generate commit message
			message, err := GenerateCommitMessage(tt.diffs)
			assert.NoError(t, err)

			// Verify message format
			assert.Contains(t, message, ":")    // Should have type and description
			assert.Contains(t, message, "\n\n") // Should have blank line before body

			// Split message into parts
			parts := strings.SplitN(message, "\n\n", 2)
			assert.Len(t, parts, 2, "Message should have two parts separated by blank line")

			// Verify first line format
			firstLine := parts[0]
			assert.Regexp(t, `^(feat|fix|docs|style|refactor|test|chore)(\([a-zA-Z0-9-]+\))?: .+$`, firstLine)

			// Verify body format
			body := parts[1]
			assert.NotEmpty(t, body, "Body should not be empty")
			assert.Contains(t, body, "\n", "Body should have multiple lines")
		})
	}
}

func TestPromptExamples(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Decode request body
		var reqBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NoError(t, err)

		// Verify prompt contains all example commit messages
		prompt := reqBody["prompt"].(string)
		assert.Contains(t, prompt, "feat(auth): add user authentication")
		assert.Contains(t, prompt, "fix(api): handle null pointer in user service")
		assert.Contains(t, prompt, "refactor(core): improve error handling")

		// Send response
		response := map[string]interface{}{
			"response": "feat: test response",
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

	// Generate commit message
	_, err = GenerateCommitMessage([]string{"test diff"})
	assert.NoError(t, err)
}
