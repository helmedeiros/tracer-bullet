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

			// Verify config file
			configFile := filepath.Join(utils.TestConfigDir, config.DefaultConfigFile)
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
	tmpDir, _, originalDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, tmpDir, originalDir)

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
			configFile := filepath.Join(utils.TestConfigDir, config.DefaultConfigFile)
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
