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

IMPORTANT INSTRUCTIONS:
1. DO NOT include any conversational messages, questions, or closing remarks
2. DO NOT ask for feedback or confirmation
3. ONLY provide the commit message in the exact format specified below
4. DO NOT add any additional text outside the commit message format
5. CAREFULLY analyze the actual code changes in the diff
6. Focus on the specific changes made to the code, not general improvements
7. Be precise about what was added, removed, or modified
8. DO NOT make assumptions about code functionality not shown in the diff
9. ONLY describe changes that are visible in the provided code diff
10. DO NOT describe changes that are not in the diff
11. DO NOT make up function names or variables that are not in the code
12. ONLY reference code that is actually shown in the changes

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
  * Be specific about which part of the codebase was changed
  * The scope should match the directory structure in the diff
  * DO NOT make up scopes that aren't in the code

- description: A clear, concise summary of the change in present tense, imperative mood
  * Good: "add user authentication"
  * Bad: "added user authentication" or "adding user authentication"
  * Must be under 50 characters
  * Must describe the main purpose of the changes
  * Must reflect the actual code changes in the diff
  * Should match the visible changes in the code
  * DO NOT describe changes that aren't in the diff

- body: REQUIRED. Must include:
  * A detailed explanation of what changed and why
  * A bullet-point list of specific changes made
  * Technical details that might be important
  * Impact of the changes
  * Any breaking changes or migration steps if applicable
  * Each bullet point should reference specific changes from the diff
  * Each change should be verifiable in the provided code diff
  * DO NOT include changes that aren't in the diff

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
10. DO NOT include any conversational text or questions
11. Each bullet point must correspond to a specific change in the diff
12. Be precise about what was added, removed, or modified
13. DO NOT describe functionality not shown in the code changes
14. Verify each change against the actual diff before including it
15. DO NOT make up function names or variables
16. ONLY reference code that is actually shown in the changes

Here is a template for your commit message based on the actual changes:

<type>(<scope>): <description>

<blank line>
<body>

For example, if the changes are about improving the prompt formatting:
feat(utils): improve prompt formatting for change analysis

Enhance how changes are presented in the prompt.
- Add clear labeling for additions, removals, and context
- Improve change visibility with better formatting
- Add explicit instructions about change analysis
- Include summary section for better overview

Here are the changes to analyze:
`

	// Add all diffs to the prompt with better structure
	prompt += "\nCHANGES TO ANALYZE:\n\n"

	// Group changes by file
	fileChanges := make(map[string][]string)
	for _, diff := range diffs {
		// Split diff into lines
		lines := strings.Split(diff, "\n")
		if len(lines) == 0 {
			continue
		}

		// Extract file name from first line
		fileLine := lines[0]
		if !strings.HasPrefix(fileLine, "File: ") {
			continue
		}
		fileName := strings.TrimPrefix(fileLine, "File: ")

		// Store the rest of the diff for this file
		fileChanges[fileName] = append(fileChanges[fileName], strings.Join(lines[1:], "\n"))
	}

	// Format changes for each file
	for fileName, changes := range fileChanges {
		prompt += fmt.Sprintf("File: %s\n", fileName)
		prompt += "Changes:\n"
		for _, change := range changes {
			// Parse the diff header (e.g., @@ -10,6 +10,7 @@)
			lines := strings.Split(change, "\n")
			if len(lines) > 0 && strings.HasPrefix(lines[0], "@@") {
				// Extract line numbers
				lineInfo := lines[0]
				prompt += fmt.Sprintf("Lines affected: %s\n", lineInfo)

				// Add the actual changes with clear labeling
				prompt += "Actual changes in this section:\n"
				for _, line := range lines[1:] {
					if strings.HasPrefix(line, "+") {
						prompt += fmt.Sprintf("  [+] Added: %s\n", strings.TrimPrefix(line, "+"))
					} else if strings.HasPrefix(line, "-") {
						prompt += fmt.Sprintf("  [-] Removed: %s\n", strings.TrimPrefix(line, "-"))
					} else {
						prompt += fmt.Sprintf("  [ ] Context: %s\n", line)
					}
				}
			}
			prompt += "\n"
		}
		prompt += "\n"
	}

	// Add a summary of the changes
	prompt += "\nSUMMARY OF CHANGES:\n"
	for fileName, changes := range fileChanges {
		prompt += fmt.Sprintf("- In %s:\n", fileName)
		for _, change := range changes {
			lines := strings.Split(change, "\n")
			if len(lines) > 0 && strings.HasPrefix(lines[0], "@@") {
				// Count additions and removals
				additions := 0
				removals := 0
				for _, line := range lines[1:] {
					if strings.HasPrefix(line, "+") {
						additions++
					} else if strings.HasPrefix(line, "-") {
						removals++
					}
				}
				prompt += fmt.Sprintf("  * %d lines added, %d lines removed\n", additions, removals)
			}
		}
	}
	prompt += "\n"

	// Add explicit instructions about the changes
	prompt += "\nIMPORTANT: When writing the commit message:\n"
	prompt += "1. ONLY describe the changes shown above\n"
	prompt += "2. DO NOT make up changes that aren't in the diff\n"
	prompt += "3. DO NOT describe functionality that isn't shown in the code\n"
	prompt += "4. Focus on the actual additions and removals shown\n"
	prompt += "5. Be specific about what was added or removed\n"
	prompt += "6. Reference the actual code changes shown above\n\n"

	// Call llama API
	reqBody := map[string]interface{}{
		"model":       "llama3",
		"prompt":      prompt,
		"max_tokens":  1000,
		"temperature": 0.7,
		"stream":      false,
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

	// Remove any conversational or closing messages
	lines := strings.Split(message, "\n")
	var cleanedLines []string
	for _, line := range lines {
		// Skip lines that look like conversational messages
		if strings.HasPrefix(strings.ToLower(line), "let me know") ||
			strings.HasPrefix(strings.ToLower(line), "i hope this") ||
			strings.HasPrefix(strings.ToLower(line), "please let me") ||
			strings.HasPrefix(strings.ToLower(line), "if you have any") ||
			strings.HasPrefix(strings.ToLower(line), "feel free to") ||
			strings.HasPrefix(strings.ToLower(line), "is there anything") ||
			strings.HasPrefix(strings.ToLower(line), "would you like") ||
			strings.HasPrefix(strings.ToLower(line), "do you need") {
			continue
		}
		cleanedLines = append(cleanedLines, line)
	}

	if len(cleanedLines) == 0 {
		return "", fmt.Errorf("empty response after cleaning conversational messages")
	}

	// Ensure the first line follows conventional commit format
	firstLine := cleanedLines[0]
	if !strings.Contains(firstLine, ":") {
		firstLine = "feat: " + firstLine
	}

	// Reconstruct the message with proper formatting
	message = firstLine
	if len(cleanedLines) > 1 {
		message += "\n\n" + strings.Join(cleanedLines[1:], "\n")
	}

	return message, nil
}
