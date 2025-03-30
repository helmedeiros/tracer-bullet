package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/helmedeiros/tracer-bullet/internal/config"
	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func setupTestEnvironment(t *testing.T) (string, string) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "tracer-test-*")
	require.NoError(t, err)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Initialize git repository
	_, err = utils.RunCommand("git", "init")
	require.NoError(t, err)

	return tmpDir, originalDir
}

func TestConfigureProject(t *testing.T) {
	tmpDir, originalDir := setupTestEnvironment(t)
	defer os.RemoveAll(tmpDir)
	defer os.Chdir(originalDir)

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

			// Verify config file
			configDir, err := utils.GetConfigDir()
			require.NoError(t, err)
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
	tmpDir, originalDir := setupTestEnvironment(t)
	defer os.RemoveAll(tmpDir)
	defer os.Chdir(originalDir)

	// First configure a project (required for user configuration)
	err := configureProject("test-project")
	require.NoError(t, err)

	tests := []struct {
		name        string
		username    string
		expectError bool
	}{
		{
			name:        "configure user with valid name",
			username:    "john.doe",
			expectError: false,
		},
		{
			name:        "configure user with empty name",
			username:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute configure user command
			err := configureUser(tt.username)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Verify git config
			username, err := utils.RunCommand("git", "config", "--local", "test-project.user")
			require.NoError(t, err)
			assert.Equal(t, tt.username, username)

			// Verify config file
			configDir, err := utils.GetConfigDir()
			require.NoError(t, err)
			configFile := filepath.Join(configDir, config.DefaultConfigFile)
			require.FileExists(t, configFile)

			// Read and verify config file contents
			data, err := os.ReadFile(configFile)
			require.NoError(t, err)
			var cfg config.Config
			err = yaml.Unmarshal(data, &cfg)
			require.NoError(t, err)
			assert.Equal(t, tt.username, cfg.AuthorName)
		})
	}
}
