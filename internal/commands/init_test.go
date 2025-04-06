package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/helmedeiros/tracer-bullet/internal/config"
	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestInitCommand(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "tracer-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set test config directory
	originalTestConfigDir := utils.TestConfigDir
	utils.TestConfigDir = tempDir
	defer func() {
		utils.TestConfigDir = originalTestConfigDir
	}()

	// Save current GitClient
	originalGitClient := utils.GitClient
	defer func() {
		utils.GitClient = originalGitClient
	}()

	// Use mock Git client
	utils.GitClient = utils.NewMockGit()
	mockGit := utils.GitClient.(*utils.MockGit)
	mockGit.GetGitRootFunc = func() (string, error) {
		return tempDir, nil
	}
	mockGit.GetConfigFunc = func(key string) (string, error) {
		switch key {
		case "user.name":
			return "Test User", nil
		case "user.email":
			return "test@example.com", nil
		default:
			return "", nil
		}
	}

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to change back to original directory: %v", err)
		}
	}()

	// Run the init command
	err = runInit(nil, nil)
	assert.NoError(t, err)

	// Check if .tracer directory was created
	tracerDir := ".tracer"
	info, err := os.Stat(tracerDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())

	// Check if config file was created and has correct structure
	configPath := filepath.Join(tracerDir, config.DefaultConfigFile)
	info, err = os.Stat(configPath)
	assert.NoError(t, err)
	assert.False(t, info.IsDir())

	// Load the config to verify its structure
	cfg, err := config.LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, config.DefaultGitBranch, cfg.GitBranch)
	assert.Equal(t, config.DefaultGitRemote, cfg.GitRemote)
	assert.Equal(t, config.DefaultStoryDir, cfg.StoryDir)
	assert.Equal(t, config.DefaultPairFile, cfg.PairFile)
	assert.Equal(t, "Test User", cfg.AuthorName)
	assert.Equal(t, "test@example.com", cfg.AuthorEmail)
	assert.Equal(t, filepath.Base(tempDir), cfg.GitRepo)

	// Running init again should not fail (idempotent)
	err = runInit(nil, nil)
	assert.NoError(t, err)
}
