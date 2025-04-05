package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/helmedeiros/tracer-bullet/internal/config"
	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestConfigureProject(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		expectError bool
	}{
		{
			name:        "configure project with valid name",
			projectName: "test-project",
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
			// Create a temporary directory for testing
			tmpDir := t.TempDir()
			repoDir := filepath.Join(tmpDir, "repo")
			err := os.MkdirAll(repoDir, 0755)
			assert.NoError(t, err)

			// Save current directory
			currentDir, err := os.Getwd()
			assert.NoError(t, err)

			// Change to test directory
			err = os.Chdir(repoDir)
			assert.NoError(t, err)
			defer func() {
				err = os.Chdir(currentDir)
				assert.NoError(t, err)
			}()

			// Initialize git repository
			_, err = utils.RunCommand("git", "init")
			assert.NoError(t, err)

			// Create a mock git client
			mockGit := utils.NewMockGit()
			utils.GitClient = mockGit

			// Configure mock behavior
			mockGit.(*utils.MockGit).SetConfigFunc = func(key, value string) error {
				if tt.expectError {
					return os.ErrPermission
				}
				return nil
			}

			// Create root command and add configure command
			rootCmd := &cobra.Command{Use: "tracer"}
			rootCmd.AddCommand(ConfigureCmd)

			// Set up the command arguments
			args := []string{"configure", "--project"}
			if tt.projectName != "" {
				args = append(args, tt.projectName)
			} else {
				args = append(args, "")
			}

			// Set the command's args
			rootCmd.SetArgs(args)

			// Execute the command
			err = rootCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "project name cannot be empty")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigureUser(t *testing.T) {
	tests := []struct {
		name        string
		userName    string
		projectName string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid name",
			userName:    "test-user",
			projectName: "test-project",
			expectError: false,
		},
		{
			name:        "empty name",
			userName:    "",
			projectName: "test-project",
			expectError: true,
			errorMsg:    "flag needs an argument: --user",
		},
		{
			name:        "without project",
			userName:    "test-user",
			projectName: "",
			expectError: true,
			errorMsg:    "project not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir := t.TempDir()
			repoDir := filepath.Join(tmpDir, "repo")
			err := os.MkdirAll(repoDir, 0755)
			assert.NoError(t, err)

			// Save current directory
			currentDir, err := os.Getwd()
			assert.NoError(t, err)

			// Change to test directory
			err = os.Chdir(repoDir)
			assert.NoError(t, err)
			defer func() {
				err = os.Chdir(currentDir)
				assert.NoError(t, err)
			}()

			// Initialize git repository
			_, err = utils.RunCommand("git", "init")
			assert.NoError(t, err)

			// Create a mock git client
			mockGit := utils.NewMockGit()
			utils.GitClient = mockGit

			// Configure mock behavior
			mockGit.(*utils.MockGit).SetConfigFunc = func(key, value string) error {
				return nil
			}
			mockGit.(*utils.MockGit).GetConfigFunc = func(key string) (string, error) {
				if key == "current.project" {
					if tt.projectName == "" {
						return "", fmt.Errorf("project not configured")
					}
					return tt.projectName, nil
				}
				return "", nil
			}
			mockGit.(*utils.MockGit).GetGitRootFunc = func() (string, error) {
				return repoDir, nil
			}

			// Create root command and add configure command
			rootCmd := &cobra.Command{Use: "tracer"}
			rootCmd.AddCommand(ConfigureCmd)

			// Configure project first if needed
			if tt.projectName != "" {
				// Create config directory
				configDir := filepath.Join(repoDir, ".tracer")
				err = os.MkdirAll(configDir, 0755)
				assert.NoError(t, err)

				// Create config file
				configFile := filepath.Join(configDir, "config.yaml")
				cfg := &config.Config{
					GitRepo:   tt.projectName,
					GitBranch: config.DefaultGitBranch,
					GitRemote: config.DefaultGitRemote,
				}
				data, err := yaml.Marshal(cfg)
				assert.NoError(t, err)
				err = os.WriteFile(configFile, data, utils.DefaultFilePerm)
				assert.NoError(t, err)

				args := []string{"configure", "--project", tt.projectName}
				rootCmd.SetArgs(args)
				err = rootCmd.Execute()
				assert.NoError(t, err)
			}

			// Set up the command arguments for user configuration
			args := []string{"configure", "--user"}
			if tt.userName != "" {
				args = append(args, tt.userName)
			}

			// Set the command's args
			rootCmd.SetArgs(args)

			// Execute the command
			err = rootCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigureShow(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, "repo")
	err := os.MkdirAll(repoDir, 0755)
	assert.NoError(t, err)

	// Save current directory
	currentDir, err := os.Getwd()
	assert.NoError(t, err)

	// Change to test directory
	err = os.Chdir(repoDir)
	assert.NoError(t, err)
	defer func() {
		err = os.Chdir(currentDir)
		assert.NoError(t, err)
	}()

	// Initialize git repository
	_, err = utils.RunCommand("git", "init")
	assert.NoError(t, err)

	// Create a mock git client
	mockGit := utils.NewMockGit()
	utils.GitClient = mockGit

	// Configure mock behavior
	mockGit.(*utils.MockGit).GetConfigFunc = func(key string) (string, error) {
		switch key {
		case "current.project":
			return "test-project", nil
		case "test-project.user":
			return "test-user", nil
		default:
			return "", nil
		}
	}
	mockGit.(*utils.MockGit).SetConfigFunc = func(key, value string) error {
		return nil
	}
	mockGit.(*utils.MockGit).GetGitRootFunc = func() (string, error) {
		return repoDir, nil
	}

	// Create config directory
	configDir := filepath.Join(repoDir, ".tracer")
	err = os.MkdirAll(configDir, 0755)
	assert.NoError(t, err)

	// Create config file
	configFile := filepath.Join(configDir, "config.yaml")
	cfg := &config.Config{
		GitRepo:     "test-project",
		GitBranch:   config.DefaultGitBranch,
		GitRemote:   config.DefaultGitRemote,
		AuthorName:  "test-user",
		JiraHost:    "https://jira.example.com",
		JiraProject: "TEST",
		JiraUser:    "jira-user",
		JiraToken:   "jira-token",
	}
	data, err := yaml.Marshal(cfg)
	assert.NoError(t, err)
	err = os.WriteFile(configFile, data, utils.DefaultFilePerm)
	assert.NoError(t, err)

	// Create root command and add configure command
	rootCmd := &cobra.Command{Use: "tracer"}
	rootCmd.AddCommand(ConfigureCmd)

	// Set up the command arguments
	args := []string{"configure", "show"}
	rootCmd.SetArgs(args)

	// Execute the command
	err = rootCmd.Execute()
	assert.NoError(t, err)
}
