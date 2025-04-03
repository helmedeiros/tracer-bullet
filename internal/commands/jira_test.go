package commands

import (
	"bytes"
	"testing"

	"github.com/helmedeiros/tracer-bullet/internal/config"
	"github.com/helmedeiros/tracer-bullet/internal/story"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJiraCommand(t *testing.T) {
	// Set up test environment
	tmpDir, _, originalDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, tmpDir, originalDir)

	t.Run("configure command", func(t *testing.T) {
		tests := []struct {
			name        string
			args        []string
			expectError bool
			validate    func(t *testing.T, cfg *config.Config)
		}{
			{
				name: "configure with all values",
				args: []string{
					"--host", "https://jira.example.com",
					"--token", "token123",
					"--project", "TEST",
					"--user", "user@example.com",
				},
				expectError: false,
				validate: func(t *testing.T, cfg *config.Config) {
					assert.Equal(t, "https://jira.example.com", cfg.JiraHost)
					assert.Equal(t, "token123", cfg.JiraToken)
					assert.Equal(t, "TEST", cfg.JiraProject)
					assert.Equal(t, "user@example.com", cfg.JiraUser)
				},
			},
			{
				name: "configure with only host",
				args: []string{
					"--host", "https://jira.example.com",
				},
				expectError: false,
				validate: func(t *testing.T, cfg *config.Config) {
					assert.Equal(t, "https://jira.example.com", cfg.JiraHost)
				},
			},
			{
				name:        "configure with empty host",
				args:        []string{},
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cmd := &cobra.Command{
					Use:   "configure",
					Short: "Configure Jira settings",
					RunE:  jiraConfigureCmd.RunE,
				}

				// Add flags
				cmd.Flags().String("host", "", "Jira host URL")
				cmd.Flags().String("token", "", "Jira API token")
				cmd.Flags().String("project", "", "Default Jira project key")
				cmd.Flags().String("user", "", "Jira username/email")

				// Create buffers for output
				var buf, errBuf bytes.Buffer
				cmd.SetOut(&buf)
				cmd.SetErr(&errBuf)

				// Set command arguments
				cmd.SetArgs(tt.args)

				// Execute command
				err := cmd.Execute()

				if tt.expectError {
					assert.Error(t, err)
					return
				}

				require.NoError(t, err)

				// Load config and validate
				if tt.validate != nil {
					cfg, err := config.LoadConfig()
					require.NoError(t, err)
					tt.validate(t, cfg)
				}
			})
		}
	})

	t.Run("link command", func(t *testing.T) {
		// Create a test story
		s, err := story.NewStory("Test Story", "Description", "test-user")
		require.NoError(t, err)
		err = s.Save()
		require.NoError(t, err)

		// Configure Jira
		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		cfg.JiraHost = "https://jira.example.com"
		cfg.JiraToken = "token123"
		cfg.JiraProject = "TEST"
		cfg.JiraUser = "user@example.com"
		err = config.SaveConfig(cfg)
		require.NoError(t, err)

		cmd := &cobra.Command{
			Use:   "link",
			Short: "Link a story to a Jira issue",
			RunE:  jiraLinkCmd.RunE,
		}

		// Add flags
		cmd.Flags().String("story", "", "Story ID")
		cmd.Flags().String("issue", "", "Jira issue ID")

		// Create buffers for output
		var buf, errBuf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&errBuf)

		// Set command arguments
		cmd.SetArgs([]string{
			"--story", s.ID,
			"--issue", "TEST-123",
		})

		// Execute command
		err = cmd.Execute()
		// Since we can't actually connect to Jira in tests, we expect an error
		assert.Error(t, err)
	})
}
