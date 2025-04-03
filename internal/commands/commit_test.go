package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/helmedeiros/tracer-bullet/internal/config"
	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommitCommand(t *testing.T) {
	tmpDir, repoDir, originalDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, tmpDir, originalDir)

	// First configure a project and user (required for commit command)
	err := configureProject("test-project")
	require.NoError(t, err)
	err = configureUser("john.doe")
	require.NoError(t, err)

	// Create a test file to commit
	testFile := filepath.Join(repoDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Initialize Git repository
	_, err = utils.RunCommand("git", "init")
	require.NoError(t, err)

	// Configure Git user for the test repository
	_, err = utils.RunCommand("git", "config", "user.name", "Test User")
	require.NoError(t, err)
	_, err = utils.RunCommand("git", "config", "user.email", "test@example.com")
	require.NoError(t, err)

	_, err = utils.RunCommand("git", "add", testFile)
	require.NoError(t, err)

	tests := []struct {
		name        string
		commitType  string
		message     string
		scope       string
		body        string
		breaking    bool
		jira        bool
		expectError bool
		expectedMsg string
	}{
		{
			name:        "basic commit",
			commitType:  "feat",
			message:     "add login functionality",
			expectError: false,
			expectedMsg: "feat: add login functionality",
		},
		{
			name:        "commit with scope",
			commitType:  "fix",
			message:     "fix login validation",
			scope:       "auth",
			expectError: false,
			expectedMsg: "fix(auth): fix login validation",
		},
		{
			name:        "commit with body",
			commitType:  "docs",
			message:     "update README",
			body:        "Added installation instructions",
			expectError: false,
			expectedMsg: "docs: update README\n\nAdded installation instructions",
		},
		{
			name:        "breaking change",
			commitType:  "feat",
			message:     "change authentication method",
			breaking:    true,
			expectError: false,
			expectedMsg: "feat!: change authentication method\n\nBREAKING CHANGE: change authentication method",
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
		{
			name:        "commit with jira story",
			commitType:  "feat",
			message:     "add user profile",
			jira:        true,
			expectError: false,
			expectedMsg: "feat: add user profile\n\nJira: https://test-jira.com/browse/TEST-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create and stage a new file for each test
			testFile := filepath.Join(repoDir, fmt.Sprintf("test_%s.txt", tt.name))
			err = os.WriteFile(testFile, []byte("test content"), 0644)
			require.NoError(t, err)
			_, err = utils.RunCommand("git", "add", testFile)
			require.NoError(t, err)

			// Set up Jira config if needed
			if tt.jira {
				cfg, err := config.LoadConfig()
				require.NoError(t, err)
				cfg.JiraHost = "test-jira.com"
				cfg.JiraProject = "TEST"
				err = config.SaveConfig(cfg)
				require.NoError(t, err)

				// Set current story in git config
				_, err = utils.RunCommand("git", "config", "--local", "TEST.current.story", "123")
				require.NoError(t, err)
			}

			// Create a new command instance for each test
			cmd := &cobra.Command{
				Use:   "create",
				Short: "Create a new commit",
				RunE:  commitCreateCmd.RunE,
			}

			// Add flags to the command
			cmd.Flags().String("type", "", "Commit type")
			cmd.Flags().String("message", "", "Commit message")
			cmd.Flags().String("scope", "", "Commit scope")
			cmd.Flags().String("body", "", "Commit body")
			cmd.Flags().Bool("breaking", false, "Breaking change")
			cmd.Flags().Bool("jira", false, "Include Jira URL")

			// Create a buffer to capture output
			var buf, errBuf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&errBuf)

			// Build command arguments
			args := []string{"--type", tt.commitType, "--message", tt.message}
			if tt.scope != "" {
				args = append(args, "--scope", tt.scope)
			}
			if tt.body != "" {
				args = append(args, "--body", tt.body)
			}
			if tt.breaking {
				args = append(args, "--breaking")
			}
			if tt.jira {
				args = append(args, "--jira")
			}

			// Set command arguments
			cmd.SetArgs(args)

			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify commit was created
			output, err := utils.RunCommand("git", "log", "-1", "--pretty=format:%B")
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, output)

			// Clean up git config if we set it
			if tt.jira {
				_, err = utils.RunCommand("git", "config", "--local", "--unset", "TEST.current.story")
				require.NoError(t, err)
			}
		})
	}
}
