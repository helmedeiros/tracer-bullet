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

IMPORTANT: You MUST analyze the actual code changes provided below and generate a message that specifically describes those changes.

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

- scope: REQUIRED. What part of the codebase is affected (e.g., api, core, ui, tests)
  * Look at the file paths in the changes to determine the scope
  * If changes affect multiple scopes, use the most relevant one

- description: A clear, concise summary of the change in present tense, imperative mood
  * Good: "add user authentication"
  * Bad: "added user authentication" or "adding user authentication"
  * Must be under 50 characters
  * Must describe the main purpose of the changes

- body: REQUIRED. Must include:
  * A detailed explanation of what changed and why
  * A bullet-point list of specific changes made
  * Technical details that might be important
  * Impact of the changes
  * Any breaking changes or migration steps if applicable

IMPORTANT RULES:
1. ALWAYS include a scope in parentheses
2. ALWAYS include a detailed body section
3. Use present tense, imperative mood for the description
4. Start each bullet point with a capital letter
5. End each bullet point with a period
6. Keep the description under 50 characters
7. Make the body comprehensive but concise
8. Focus on the actual changes shown in the diff
9. Include specific file names and line numbers when relevant

Examples of good commit messages:

1. Feature Addition:
feat(auth): implement JWT authentication

Add JWT-based authentication system to secure API endpoints.
- Implement JWT token generation and validation
- Add middleware for protected routes
- Update API documentation with authentication details
- Add test coverage for authentication flow

2. Bug Fix:
fix(api): handle null pointer in user service

Prevent application crash when user data is missing.
- Add null checks in user service methods
- Return appropriate error responses
- Add test cases for null scenarios
- Update error logging to include context

3. Refactoring:
refactor(core): improve error handling

Standardize error handling across the application.
- Create custom error types for better error classification
- Implement consistent error response format
- Update error logging with structured data
- Add documentation for error handling patterns

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

	// Debug: Print the full response
	fmt.Printf("Llama API Response: %+v\n", result)

	// Try different response field names
	var message string
	if response, ok := result["response"].(string); ok {
		message = response
	} else if response, ok := result["text"].(string); ok {
		message = response
	} else if response, ok := result["content"].(string); ok {
		message = response
	} else {
		return "", fmt.Errorf("no valid response field found in llama API response")
	}

	// Clean up the response
	message = strings.TrimSpace(message)

	// Split the message into lines
	lines := strings.Split(message, "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("empty response from llama API")
	}

	// Ensure the first line follows conventional commit format
	firstLine := lines[0]
	if !strings.Contains(firstLine, ":") {
		firstLine = "feat: " + firstLine
	}

	// Reconstruct the message with proper formatting
	message = firstLine
	if len(lines) > 1 {
		message += "\n\n" + strings.Join(lines[1:], "\n")
	}

	return message, nil
}
