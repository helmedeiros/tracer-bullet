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
	prompt := buildPrompt(diffs)
	return callLlamaAPI(prompt)
}

func buildPrompt(diffs []string) string {
	prompt := getBasePrompt()
	prompt += "\nCHANGES TO ANALYZE:\n\n"

	fileChanges := groupChangesByFile(diffs)
	prompt += formatFileChanges(fileChanges)
	prompt += formatChangeSummary(fileChanges)

	return prompt
}

func getBasePrompt() string {
	return `You are an expert at writing clear and descriptive commit messages.
The commit message MUST follow this exact format:
<type>(<scope>): <description>

- type: feat, fix, docs, style, refactor, test, or chore
- scope: optional component name in parentheses
- description: start with verb, use imperative mood, no period

A blank line must separate the header from the body.
The body should list the key changes with bullet points.

Examples of good commit messages:

feat(auth): add user authentication

Implement user authentication flow with secure password handling
- Add login/logout endpoints
- Create password hashing utilities
- Add session management
- Implement JWT token generation

fix(api): handle null pointer in user service

Update user lookup to handle missing profiles
- Add null checks in user service
- Improve error messages
- Add validation for user IDs

refactor(core): improve error handling

Standardize error handling across core services
- Create custom error types
- Add error wrapping
- Improve error messages
- Add error logging

The changes will be provided in git diff format. Generate a commit message for these changes:`
}

func groupChangesByFile(diffs []string) map[string][]string {
	fileChanges := make(map[string][]string)
	for _, diff := range diffs {
		lines := strings.Split(diff, "\n")
		if len(lines) == 0 {
			continue
		}

		fileLine := lines[0]
		if !strings.HasPrefix(fileLine, "File: ") {
			continue
		}
		fileName := strings.TrimPrefix(fileLine, "File: ")
		fileChanges[fileName] = append(fileChanges[fileName], strings.Join(lines[1:], "\n"))
	}
	return fileChanges
}

func formatFileChanges(fileChanges map[string][]string) string {
	var prompt string
	for fileName, changes := range fileChanges {
		prompt += fmt.Sprintf("File: %s\n", fileName)
		for _, change := range changes {
			lines := strings.Split(change, "\n")
			if len(lines) > 0 && strings.HasPrefix(lines[0], "@@") {
				prompt += lines[0] + "\n"
				for _, line := range lines[1:] {
					prompt += line + "\n"
				}
			}
		}
		prompt += "\n"
	}
	return prompt
}

func formatChangeSummary(fileChanges map[string][]string) string {
	summary := "\nSUMMARY OF CHANGES:\n"
	for fileName, changes := range fileChanges {
		summary += fmt.Sprintf("- In %s:\n", fileName)
		for _, change := range changes {
			summary += formatChangeStats(change)
		}
	}
	return summary
}

func formatChangeStats(change string) string {
	lines := strings.Split(change, "\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "@@") {
		return ""
	}

	additions := 0
	removals := 0
	for _, line := range lines[1:] {
		switch {
		case strings.HasPrefix(line, "+"):
			additions++
		case strings.HasPrefix(line, "-"):
			removals++
		}
	}

	return fmt.Sprintf("  - %d lines added, %d lines removed\n", additions, removals)
}

func callLlamaAPI(prompt string) (string, error) {
	response, err := makeAPIRequest(prompt)
	if err != nil {
		return "", err
	}

	message, err := extractMessageFromResponse(response)
	if err != nil {
		return "", err
	}

	return cleanAndFormatMessage(message)
}

func makeAPIRequest(prompt string) (map[string]interface{}, error) {
	reqBody := map[string]interface{}{
		"model":       "llama3",
		"prompt":      prompt,
		"max_tokens":  500,
		"temperature": 0.7,
		"stream":      false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	if err := validateAPIURL(); err != nil {
		return nil, err
	}

	//nolint:gosec // URL is validated before making the request
	resp, err := http.Post(llamaAPIURL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to call llama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("llama API returned status code %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode llama API response: %w", err)
	}

	return result, nil
}

func validateAPIURL() error {
	_, err := url.ParseRequestURI(llamaAPIURL)
	if err != nil {
		return fmt.Errorf("invalid llama API URL: %w", err)
	}
	return nil
}

func extractMessageFromResponse(result map[string]interface{}) (string, error) {
	responseFields := []string{"response", "text", "content"}

	for _, field := range responseFields {
		if response, ok := result[field].(string); ok {
			return strings.TrimSpace(response), nil
		}
	}

	return "", fmt.Errorf("missing response field in llama API response")
}

func cleanAndFormatMessage(message string) (string, error) {
	lines := strings.Split(message, "\n")
	cleanedLines := filterConversationalLines(lines)

	if len(cleanedLines) == 0 {
		return "", fmt.Errorf("empty response after cleaning conversational messages")
	}

	return formatCommitMessage(cleanedLines), nil
}

func filterConversationalLines(lines []string) []string {
	conversationalPrefixes := []string{
		"let me know",
		"i hope this",
		"please let me",
		"if you have any",
		"feel free to",
		"is there anything",
		"would you like",
		"do you need",
	}

	var cleanedLines []string
	for _, line := range lines {
		if !hasConversationalPrefix(line, conversationalPrefixes) {
			cleanedLines = append(cleanedLines, line)
		}
	}
	return cleanedLines
}

func hasConversationalPrefix(line string, prefixes []string) bool {
	lowercaseLine := strings.ToLower(line)
	for _, prefix := range prefixes {
		if strings.HasPrefix(lowercaseLine, prefix) {
			return true
		}
	}
	return false
}

func formatCommitMessage(lines []string) string {
	firstLine := lines[0]
	if !strings.Contains(firstLine, ":") {
		firstLine = "feat: " + firstLine
	}

	if len(lines) == 1 {
		return firstLine
	}

	// Remove any extra blank lines between header and body
	var bodyLines []string
	foundNonEmpty := false
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) != "" {
			foundNonEmpty = true
			bodyLines = append(bodyLines, line)
		} else if foundNonEmpty {
			bodyLines = append(bodyLines, line)
		}
	}

	return firstLine + "\n\n" + strings.Join(bodyLines, "\n")
}
