package commands

import (
	"bytes"
	"os"
	"testing"

	"github.com/helmedeiros/tracer-bullet/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJiraCommand(t *testing.T) {
	tmpDir, originalDir := setupTestEnvironment(t)
	defer func() {
		err := os.Chdir(originalDir)
		require.NoError(t, err)
		os.RemoveAll(tmpDir)
	}()

	// First configure a project (required for config)
	err := configureProject("test-project")
	require.NoError(t, err)

	tests := []struct {
		name        string
		host        string
		token       string
		project     string
		user        string
		expectError bool
	}{
		{
			name:        "configure jira with all values",
			host:        "https://mycompany.atlassian.net",
			token:       "my-secret-token",
			project:     "PROJ",
			user:        "user@example.com",
			expectError: false,
		},
		{
			name:        "configure jira with only host",
			host:        "https://another.atlassian.net",
			expectError: false,
		},
		{
			name:        "configure jira with empty host",
			host:        "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new command instance for each test
			cmd := &cobra.Command{
				Use:   "configure",
				Short: "Configure Jira settings",
				Long:  `Configure Jira settings including host, project, and authentication.`,
				RunE:  jiraConfigureCmd.RunE,
			}

			// Add flags to the command
			cmd.Flags().String("host", "", "Jira host URL")
			cmd.Flags().String("token", "", "Jira API token")
			cmd.Flags().String("project", "", "Default Jira project key")
			cmd.Flags().String("user", "", "Jira username/email")

			// Create a buffer to capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)

			// Build command arguments
			var args []string
			if tt.host != "" {
				args = append(args, "--host", tt.host)
			}
			if tt.token != "" {
				args = append(args, "--token", tt.token)
			}
			if tt.project != "" {
				args = append(args, "--project", tt.project)
			}
			if tt.user != "" {
				args = append(args, "--user", tt.user)
			}

			// Set command arguments
			cmd.SetArgs(args)

			// Execute command
			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Load config and verify values
			cfg, err := config.LoadConfig()
			require.NoError(t, err)

			if tt.host != "" {
				assert.Equal(t, tt.host, cfg.JiraHost, "Jira host should match")
			}
			if tt.token != "" {
				assert.Equal(t, tt.token, cfg.JiraToken, "Jira token should match")
			}
			if tt.project != "" {
				assert.Equal(t, tt.project, cfg.JiraProject, "Jira project should match")
			}
			if tt.user != "" {
				assert.Equal(t, tt.user, cfg.JiraUser, "Jira user should match")
			}

			// Verify output format
			output := buf.String()
			assert.Contains(t, output, "Jira configuration updated:", "Output should contain update message")
			assert.Contains(t, output, "Host:", "Output should show host")
			assert.Contains(t, output, "Project:", "Output should show project")
			assert.Contains(t, output, "User:", "Output should show user")
			assert.Contains(t, output, "Token:", "Output should show token status")
		})
	}
}
