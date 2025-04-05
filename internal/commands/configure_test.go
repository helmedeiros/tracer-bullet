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
	"gopkg.in/yaml.v3"
)

func TestConfigureProject(t *testing.T) {
	tmpDir, _, originalDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, tmpDir, originalDir)

	tests := []struct {
		name           string
		projectName    string
		expectedConfig *config.Config
		expectError    bool
	}{
		{
			name:        "configure project with valid name",
			projectName: "test-project",
			expectedConfig: &config.Config{
				GitRepo:   "test-project",
				GitBranch: config.DefaultGitBranch,
				GitRemote: config.DefaultGitRemote,
			},
			expectError: false,
		},
		{
			name:        "configure project with empty name",
			projectName: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute configure project command
			err := configureProject(tt.projectName)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Verify git config
			projectName, err := utils.RunCommand("git", "config", "--local", "current.project")
			require.NoError(t, err)
			assert.Equal(t, tt.projectName, projectName)

			// Get repository-specific config directory
			configDir, err := utils.GetRepoConfigDir()
			require.NoError(t, err)

			// Verify config file
			configFile := filepath.Join(configDir, config.DefaultConfigFile)
			require.FileExists(t, configFile)

			// Read and verify config file contents
			data, err := os.ReadFile(configFile)
			require.NoError(t, err)
			var cfg config.Config
			err = yaml.Unmarshal(data, &cfg)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedConfig.GitRepo, cfg.GitRepo)
			assert.Equal(t, tt.expectedConfig.GitBranch, cfg.GitBranch)
			assert.Equal(t, tt.expectedConfig.GitRemote, cfg.GitRemote)
		})
	}
}

func TestConfigureUser(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		expectError bool
		errorMsg    string
		setup       func() error
	}{
		{
			name:        "valid_name",
			username:    "john.doe",
			expectError: false,
			setup: func() error {
				return configureProject("test-project")
			},
		},
		{
			name:        "empty_name",
			username:    "",
			expectError: true,
			errorMsg:    "username cannot be empty",
			setup: func() error {
				return configureProject("test-project")
			},
		},
		{
			name:        "without_project",
			username:    "john.doe",
			expectError: true,
			errorMsg:    "project not configured. Please run 'tracer configure project' first",
			setup: func() error {
				return nil // No setup needed, we want to test without project configuration
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test environment
			tmpDir, _, originalDir := setupTestEnvironment(t)
			defer cleanupTestEnvironment(t, tmpDir, originalDir)

			// Run setup if provided
			if tt.setup != nil {
				err := tt.setup()
				require.NoError(t, err)
			}

			// Execute configure user command
			err := configureUser(tt.username)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Equal(t, tt.errorMsg, err.Error())
				}
				return
			}
			require.NoError(t, err)

			// Only verify git config and config file contents if we don't expect an error
			if !tt.expectError {
				// Verify git config
				projectName, err := utils.RunCommand("git", "config", "--local", "current.project")
				require.NoError(t, err)
				username, err := utils.RunCommand("git", "config", "--local", fmt.Sprintf("%s.user", projectName))
				require.NoError(t, err)
				assert.Equal(t, tt.username, username)

				// Get repository-specific config directory
				configDir, err := utils.GetRepoConfigDir()
				require.NoError(t, err)

				// Verify config file
				configFile := filepath.Join(configDir, config.DefaultConfigFile)
				require.FileExists(t, configFile)

				// Read and verify config file contents
				data, err := os.ReadFile(configFile)
				require.NoError(t, err)
				var cfg config.Config
				err = yaml.Unmarshal(data, &cfg)
				require.NoError(t, err)
				assert.Equal(t, tt.username, cfg.AuthorName)
			}
		})
	}
}

func TestConfigureShow(t *testing.T) {
	tmpDir, _, originalDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, tmpDir, originalDir)

	// First configure a project and user
	err := configureProject("test-project")
	require.NoError(t, err)
	err = configureUser("john.doe")
	require.NoError(t, err)

	// Create a test command
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE:  configureShowCmd.RunE,
	}

	// Create a buffer to capture output
	var buf, errBuf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&errBuf)

	// Execute the command
	err = cmd.Execute()
	require.NoError(t, err)

	// Verify output
	output := buf.String()
	assert.Contains(t, output, "Current Configuration:")
	assert.Contains(t, output, "Project: test-project")
	assert.Contains(t, output, "User: john.doe")
	assert.Contains(t, output, "Jira:")
	assert.Contains(t, output, "  Host: ")
	assert.Contains(t, output, "  Project: ")
	assert.Contains(t, output, "  User: ")
	assert.Contains(t, output, "  Token: [NOT CONFIGURED]")
}
