package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommitCommand(t *testing.T) {
	tmpDir, originalDir := setupTestEnvironment(t)
	defer func() {
		err := os.Chdir(originalDir)
		require.NoError(t, err)
		os.RemoveAll(tmpDir)
	}()

	// Initialize git repo
	err := utils.RunGitInit()
	require.NoError(t, err)

	// Create and configure test repository
	err = configureProject("test-project")
	require.NoError(t, err)
	err = configureUser("test.user")
	require.NoError(t, err)

	// Create a test file and stage it
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	_, err = utils.RunCommand("git", "add", "test.txt")
	require.NoError(t, err)

	tests := []struct {
		name        string
		commitType  string
		scope       string
		message     string
		body        string
		breaking    bool
		expectError bool
	}{
		{
			name:        "create feat commit",
			commitType:  "feat",
			message:     "add new feature",
			expectError: false,
		},
		{
			name:        "create feat commit with scope",
			commitType:  "feat",
			scope:       "auth",
			message:     "add login functionality",
			expectError: false,
		},
		{
			name:        "create fix commit with body",
			commitType:  "fix",
			message:     "fix login issue",
			body:        "This fixes a critical issue with the login system\nWhere users could not log in with correct credentials",
			expectError: false,
		},
		{
			name:        "create breaking change",
			commitType:  "feat",
			scope:       "api",
			message:     "change authentication method",
			body:        "Switch from basic auth to OAuth2",
			breaking:    true,
			expectError: false,
		},
		{
			name:        "invalid commit type",
			commitType:  "invalid",
			message:     "test message",
			expectError: true,
		},
		{
			name:        "empty message",
			commitType:  "feat",
			message:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset test file for each test
			err := os.WriteFile(testFile, []byte("test content "+tt.name), 0644)
			require.NoError(t, err)
			_, err = utils.RunCommand("git", "add", "test.txt")
			require.NoError(t, err)

			// Create a new command instance for each test
			cmd := &cobra.Command{
				Use:   "create",
				Short: "Create a new commit",
				Long:  `Create a new git commit with proper conventional commit format.`,
				RunE:  commitCreateCmd.RunE,
			}

			// Add flags to the command
			cmd.Flags().String("type", "", "Commit type")
			cmd.Flags().String("scope", "", "Commit scope")
			cmd.Flags().String("message", "", "Commit message")
			cmd.Flags().String("body", "", "Commit body")
			cmd.Flags().Bool("breaking", false, "Mark as breaking change")
			cmd.MarkFlagRequired("type")
			cmd.MarkFlagRequired("message")

			// Create a buffer to capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)

			// Build command arguments
			var args []string
			args = append(args, "--type", tt.commitType)
			if tt.scope != "" {
				args = append(args, "--scope", tt.scope)
			}
			args = append(args, "--message", tt.message)
			if tt.body != "" {
				args = append(args, "--body", tt.body)
			}
			if tt.breaking {
				args = append(args, "--breaking")
			}

			// Set command arguments
			cmd.SetArgs(args)

			// Execute command
			err = cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify commit was created with correct message
			output := buf.String()
			assert.Contains(t, output, "Created commit:", "Output should confirm commit creation")
			assert.Contains(t, output, tt.commitType, "Output should contain commit type")
			if tt.scope != "" {
				assert.Contains(t, output, "("+tt.scope+")", "Output should contain scope")
			}
			assert.Contains(t, output, tt.message, "Output should contain commit message")

			// Verify git log
			log, err := utils.RunCommand("git", "log", "-1", "--pretty=format:%B")
			require.NoError(t, err)

			// Verify commit message format
			expectedMsg := tt.commitType
			if tt.scope != "" {
				expectedMsg += "(" + tt.scope + ")"
			}
			if tt.breaking {
				expectedMsg += "!"
			}
			expectedMsg += ": " + tt.message

			assert.True(t, strings.HasPrefix(log, expectedMsg), "Git log should match expected commit message format")

			if tt.body != "" {
				assert.Contains(t, log, tt.body, "Git log should contain commit body")
			}

			if tt.breaking {
				assert.Contains(t, log, "BREAKING CHANGE:", "Git log should contain breaking change footer")
			}
		})
	}
}
