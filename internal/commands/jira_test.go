package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/helmedeiros/tracer-bullet/internal/config"
	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestJiraCommand(t *testing.T) {
	tmpDir, _, originalDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, tmpDir, originalDir)

	// First configure a project and user (required for jira configuration)
	err := configureProject("test-project")
	require.NoError(t, err)
	err = configureUser("john.doe")
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
			host:        "https://jira.example.com",
			token:       "test-token",
			project:     "TEST",
			user:        "test.user",
			expectError: false,
		},
		{
			name:        "configure jira with only host",
			host:        "https://jira.example.com",
			expectError: false,
		},
		{
			name:        "configure jira with empty host",
			host:        "",
			expectError: true,
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
			err = cmd.Execute()
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Verify config file
			configFile := filepath.Join(utils.TestConfigDir, config.DefaultConfigFile)
			require.FileExists(t, configFile)

			// Read and verify config file contents
			data, err := os.ReadFile(configFile)
			require.NoError(t, err)
			var cfg config.Config
			err = yaml.Unmarshal(data, &cfg)
			require.NoError(t, err)
			assert.Equal(t, tt.host, cfg.JiraHost)
			if tt.token != "" {
				assert.Equal(t, tt.token, cfg.JiraToken)
			}
			if tt.project != "" {
				assert.Equal(t, tt.project, cfg.JiraProject)
			}
			if tt.user != "" {
				assert.Equal(t, tt.user, cfg.JiraUser)
			}
		})
	}
}
