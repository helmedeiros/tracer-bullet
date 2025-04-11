package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// llamaAPIURL is the URL of the llama API server
var llamaAPIURL = "http://localhost:11434/api/generate"

// SetLlamaAPIURL sets the URL of the llama API server
func SetLlamaAPIURL(urlStr string) error {
	// Validate the URL
	_, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	llamaAPIURL = urlStr
	return nil
}

// GenerateCommitMessage uses llama to generate a commit message from the provided diffs
func GenerateCommitMessage(diffs []string) (string, error) {
	// Prepare the prompt for llama
	prompt := `You are an expert at writing clear and descriptive commit messages following the conventional commit format.
Please analyze the following code changes and generate a commit message that clearly explains what changed and why.

The commit message MUST follow this exact format:
<type>(<scope>): <description>

<blank line>

<body>

Where:
- type: One of these exactly:
  * feat: A new feature
  * fix: A bug fix
  * docs: Documentation changes
  * style: Code style changes (formatting, etc.)
  * refactor: Code refactoring (no functional changes)
  * test: Adding or modifying tests
  * chore: Maintenance tasks, build process, etc.

- scope: Optional, but recommended. What part of the codebase is affected (e.g., api, core, ui, tests)

- description: A clear, concise summary of the change in present tense, imperative mood
  * Good: "add user authentication"
  * Bad: "added user authentication" or "adding user authentication"

- body: Detailed explanation of:
  * What changed and why
  * Any technical details that might be important
  * Impact of the changes
  * Any breaking changes or migration steps if applicable

Examples of good commit messages:

1. Simple feature:
feat(auth): add user authentication

Add JWT-based authentication system to secure API endpoints.
- Implement JWT token generation and validation
- Add middleware for protected routes
- Update API documentation

2. Bug fix:
fix(api): handle null pointer in user service

Prevent application crash when user data is missing.
- Add null checks in user service methods
- Return appropriate error responses
- Add test cases for null scenarios

3. Refactoring:
refactor(core): improve error handling

Standardize error handling across the application.
- Create custom error types
- Implement consistent error responses
- Update error logging

Here are the changes to analyze:
`

	// Add all diffs to the prompt
	for _, diff := range diffs {
		prompt += "\n" + diff + "\n"
	}

	// Call llama API
	reqBody := map[string]interface{}{
		"model":       "llama3",
		"prompt":      prompt,
		"max_tokens":  500,
		"temperature": 0.7,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Validate the URL before making the request
	_, err = url.ParseRequestURI(llamaAPIURL)
	if err != nil {
		return "", fmt.Errorf("invalid llama API URL: %w", err)
	}

	// Make the HTTP request to the Llama API
	//nolint:gosec // URL is validated before making the request
	resp, err := http.Post(llamaAPIURL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to call llama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llama API returned status code %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode llama API response: %w", err)
	}

	// Check if response field exists
	response, ok := result["response"]
	if !ok {
		return "", fmt.Errorf("missing response field in llama API response")
	}

	// Convert response to string
	message, ok := response.(string)
	if !ok {
		return "", fmt.Errorf("invalid response field type in llama API response")
	}

	// Clean up the response
	message = strings.TrimSpace(message)

	// Ensure the message follows conventional commit format
	if !strings.Contains(message, ":") {
		// If no type is specified, default to feat
		message = "feat: " + message
	}

	return message, nil
}
