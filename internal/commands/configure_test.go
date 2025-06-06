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

func TestConfigureClean(t *testing.T) {
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

	// Create a mock git client
	mockGit := utils.NewMockGit()
	utils.GitClient = mockGit

	// Configure mock behavior
	mockGit.(*utils.MockGit).GetGitRootFunc = func() (string, error) {
		return repoDir, nil
	}

	// Create root command and add configure command
	rootCmd := &cobra.Command{Use: "tracer"}
	rootCmd.AddCommand(ConfigureCmd)

	// Set up the command arguments
	args := []string{"configure", "clean"}
	rootCmd.SetArgs(args)

	// Create a buffer to capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	// Execute the command
	err = rootCmd.Execute()
	assert.NoError(t, err)

	// Verify help output
	output := buf.String()
	assert.Contains(t, output, "Remove tracer configurations")
	assert.Contains(t, output, "Use subcommands to clean specific configurations")
	assert.Contains(t, output, "all")
	assert.Contains(t, output, "git")
	assert.Contains(t, output, "jira")
	assert.Contains(t, output, "stories")
}

func TestConfigureCleanGit(t *testing.T) {
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

	// Create a mock git client
	mockGit := utils.NewMockGit()
	utils.GitClient = mockGit

	// Configure mock behavior
	mockGit.(*utils.MockGit).SetConfigFunc = func(key, value string) error {
		return nil
	}
	mockGit.(*utils.MockGit).GetConfigFunc = func(key string) (string, error) {
		if key == "current.project" {
			return "test-project", nil
		}
		return "", nil
	}
	mockGit.(*utils.MockGit).GetGitRootFunc = func() (string, error) {
		return repoDir, nil
	}

	// Create root command and add configure command
	rootCmd := &cobra.Command{Use: "tracer"}
	rootCmd.AddCommand(ConfigureCmd)

	// Set up the command arguments
	args := []string{"configure", "clean", "git"}
	rootCmd.SetArgs(args)

	// Create a buffer to capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	// Execute the command
	err = rootCmd.Execute()
	assert.NoError(t, err)

	// Verify output
	output := buf.String()
	assert.Contains(t, output, "Git configurations have been removed")
}

func TestConfigureCleanStories(t *testing.T) {
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

	// Create a mock git client
	mockGit := utils.NewMockGit()
	utils.GitClient = mockGit

	// Configure mock behavior
	mockGit.(*utils.MockGit).GetGitRootFunc = func() (string, error) {
		return repoDir, nil
	}

	// Create repository-specific config directory and stories directory
	repoConfigDir := filepath.Join(repoDir, ".tracer")
	err = os.MkdirAll(repoConfigDir, 0755)
	assert.NoError(t, err)

	repoStoriesDir := filepath.Join(repoConfigDir, "stories")
	err = os.MkdirAll(repoStoriesDir, 0755)
	assert.NoError(t, err)

	// Create global config directory and stories directory
	globalConfigDir := filepath.Join(tmpDir, ".tracer")
	err = os.MkdirAll(globalConfigDir, 0755)
	assert.NoError(t, err)

	globalStoriesDir := filepath.Join(globalConfigDir, "stories")
	err = os.MkdirAll(globalStoriesDir, 0755)
	assert.NoError(t, err)

	// Override the global config directory for testing
	utils.TestConfigDir = globalConfigDir

	// Create root command and add configure command
	rootCmd := &cobra.Command{Use: "tracer"}
	rootCmd.AddCommand(ConfigureCmd)

	// Set up the command arguments
	args := []string{"configure", "clean", "stories"}
	rootCmd.SetArgs(args)

	// Create a buffer to capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	// Execute the command
	err = rootCmd.Execute()
	assert.NoError(t, err)

	// Verify output
	output := buf.String()
	assert.Contains(t, output, "Story configurations have been removed")

	// Verify that both stories directories were removed
	_, err = os.Stat(repoStoriesDir)
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(globalStoriesDir)
	assert.True(t, os.IsNotExist(err))

	// Reset test config directory
	utils.TestConfigDir = ""
}

func TestConfigureCleanJira(t *testing.T) {
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

	// Create a mock git client
	mockGit := utils.NewMockGit()
	utils.GitClient = mockGit

	// Configure mock behavior
	mockGit.(*utils.MockGit).GetGitRootFunc = func() (string, error) {
		return repoDir, nil
	}

	// Create config directory and file
	configDir := filepath.Join(repoDir, ".tracer")
	err = os.MkdirAll(configDir, 0755)
	assert.NoError(t, err)

	configFile := filepath.Join(configDir, "config.yaml")
	cfg := &config.Config{
		JiraHost:    "https://jira.example.com",
		JiraToken:   "test-token",
		JiraProject: "TEST",
		JiraUser:    "test-user",
	}
	data, err := yaml.Marshal(cfg)
	assert.NoError(t, err)
	err = os.WriteFile(configFile, data, utils.DefaultFilePerm)
	assert.NoError(t, err)

	// Create root command and add configure command
	rootCmd := &cobra.Command{Use: "tracer"}
	rootCmd.AddCommand(ConfigureCmd)

	// Set up the command arguments
	args := []string{"configure", "clean", "jira"}
	rootCmd.SetArgs(args)

	// Create a buffer to capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	// Execute the command
	err = rootCmd.Execute()
	assert.NoError(t, err)

	// Verify output
	output := buf.String()
	assert.Contains(t, output, "Jira configurations have been removed")

	// Verify that Jira settings were cleared
	updatedCfg, err := config.LoadConfig()
	assert.NoError(t, err)
	assert.Empty(t, updatedCfg.JiraHost)
	assert.Empty(t, updatedCfg.JiraToken)
	assert.Empty(t, updatedCfg.JiraProject)
	assert.Empty(t, updatedCfg.JiraUser)
}

func TestConfigureCleanAll(t *testing.T) {
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

	// Create a mock git client
	mockGit := utils.NewMockGit()
	utils.GitClient = mockGit

	// Configure mock behavior
	mockGit.(*utils.MockGit).SetConfigFunc = func(key, value string) error {
		return nil
	}
	mockGit.(*utils.MockGit).GetConfigFunc = func(key string) (string, error) {
		if key == "current.project" {
			return "test-project", nil
		}
		return "", nil
	}
	mockGit.(*utils.MockGit).GetGitRootFunc = func() (string, error) {
		return repoDir, nil
	}

	// Create repository-specific config directory and files
	repoConfigDir := filepath.Join(repoDir, ".tracer")
	err = os.MkdirAll(repoConfigDir, 0755)
	assert.NoError(t, err)

	repoConfigFile := filepath.Join(repoConfigDir, "config.yaml")
	repoCfg := &config.Config{
		GitRepo:   "test-project",
		GitBranch: config.DefaultGitBranch,
		GitRemote: config.DefaultGitRemote,
	}
	data, err := yaml.Marshal(repoCfg)
	assert.NoError(t, err)
	err = os.WriteFile(repoConfigFile, data, utils.DefaultFilePerm)
	assert.NoError(t, err)

	repoStoriesDir := filepath.Join(repoConfigDir, "stories")
	err = os.MkdirAll(repoStoriesDir, 0755)
	assert.NoError(t, err)

	// Create global config directory and files
	globalConfigDir := filepath.Join(tmpDir, ".tracer")
	err = os.MkdirAll(globalConfigDir, 0755)
	assert.NoError(t, err)

	globalConfigFile := filepath.Join(globalConfigDir, "config.yaml")
	globalCfg := &config.Config{
		GitRepo:   "global-project",
		GitBranch: config.DefaultGitBranch,
		GitRemote: config.DefaultGitRemote,
	}
	data, err = yaml.Marshal(globalCfg)
	assert.NoError(t, err)
	err = os.WriteFile(globalConfigFile, data, utils.DefaultFilePerm)
	assert.NoError(t, err)

	globalStoriesDir := filepath.Join(globalConfigDir, "stories")
	err = os.MkdirAll(globalStoriesDir, 0755)
	assert.NoError(t, err)

	// Override the global config directory for testing
	utils.TestConfigDir = globalConfigDir

	// Create root command and add configure command
	rootCmd := &cobra.Command{Use: "tracer"}
	rootCmd.AddCommand(ConfigureCmd)

	// Set up the command arguments
	args := []string{"configure", "clean", "all"}
	rootCmd.SetArgs(args)

	// Create a buffer to capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	// Execute the command
	err = rootCmd.Execute()
	assert.NoError(t, err)

	// Verify output
	output := buf.String()
	assert.Contains(t, output, "All configurations have been removed")

	// Verify that all configurations were removed
	_, err = os.Stat(repoConfigFile)
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(repoStoriesDir)
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(globalConfigFile)
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(globalStoriesDir)
	assert.True(t, os.IsNotExist(err))

	// Reset test config directory
	utils.TestConfigDir = ""
}
