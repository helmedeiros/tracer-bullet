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
	prompt := `Please analyze the following code changes and generate a commit message following the conventional commit format.
The first line should be the type and description, followed by a blank line and then a detailed description of the changes.

The commit message should follow this format:
<type>(<scope>): <description>

<blank line>

<body>

Where:
- type: feat, fix, docs, style, refactor, test, chore
- scope: optional, what part of the codebase is affected
- description: short description of the change
- body: detailed description of what and why the change was made

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
