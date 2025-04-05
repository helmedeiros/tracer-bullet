package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, DefaultGitBranch, cfg.GitBranch)
	assert.Equal(t, DefaultGitRemote, cfg.GitRemote)
	assert.Equal(t, DefaultStoryDir, cfg.StoryDir)
	assert.Equal(t, DefaultPairFile, cfg.PairFile)
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		wantErr  bool
		setup    func(t *testing.T) string
		teardown func(t *testing.T, path string)
	}{
		{
			name: "load existing config",
			config: &Config{
				GitRepo:     "test-repo",
				GitBranch:   "main",
				GitRemote:   "origin",
				StoryDir:    "stories",
				PairFile:    "pair.json",
				AuthorName:  "Test User",
				AuthorEmail: "test@example.com",
				JiraHost:    "https://jira.example.com",
				JiraToken:   "token123",
				JiraProject: "TEST",
				JiraUser:    "test@example.com",
			},
			setup: func(t *testing.T) string {
				dir := t.TempDir()

				// Save current directory
				currentDir, err := os.Getwd()
				assert.NoError(t, err)

				// Change to test directory
				err = os.Chdir(dir)
				assert.NoError(t, err)

				// Create .tracer directory
				tracerDir := filepath.Join(dir, ".tracer")
				err = os.MkdirAll(tracerDir, 0755)
				assert.NoError(t, err)

				// Change back to original directory
				err = os.Chdir(currentDir)
				assert.NoError(t, err)

				utils.TestConfigDir = tracerDir
				return dir
			},
			teardown: func(t *testing.T, path string) {
				utils.TestConfigDir = ""
			},
		},
		{
			name: "load non-existent config",
			setup: func(t *testing.T) string {
				dir := t.TempDir()

				// Save current directory
				currentDir, err := os.Getwd()
				assert.NoError(t, err)

				// Change to test directory
				err = os.Chdir(dir)
				assert.NoError(t, err)

				// Create .tracer directory
				tracerDir := filepath.Join(dir, ".tracer")
				err = os.MkdirAll(tracerDir, 0755)
				assert.NoError(t, err)

				// Change back to original directory
				err = os.Chdir(currentDir)
				assert.NoError(t, err)

				utils.TestConfigDir = tracerDir
				return dir
			},
			teardown: func(t *testing.T, path string) {
				utils.TestConfigDir = ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			defer tt.teardown(t, dir)

			// Save current directory
			currentDir, err := os.Getwd()
			assert.NoError(t, err)

			if tt.config != nil {
				// Write config file
				configFile := filepath.Join(dir, ".tracer", DefaultConfigFile)
				data, err := yaml.Marshal(tt.config)
				assert.NoError(t, err)
				err = os.WriteFile(configFile, data, 0644)
				assert.NoError(t, err)

				// Change to the Git repository directory before loading config
				err = os.Chdir(dir)
				assert.NoError(t, err)
				defer func() {
					err = os.Chdir(currentDir)
					assert.NoError(t, err)
				}()
			}

			cfg, err := LoadConfig()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, cfg)

			if tt.config != nil {
				assert.Equal(t, tt.config.GitRepo, cfg.GitRepo)
				assert.Equal(t, tt.config.GitBranch, cfg.GitBranch)
				assert.Equal(t, tt.config.GitRemote, cfg.GitRemote)
				assert.Equal(t, tt.config.StoryDir, cfg.StoryDir)
				assert.Equal(t, tt.config.PairFile, cfg.PairFile)
				assert.Equal(t, tt.config.AuthorName, cfg.AuthorName)
				assert.Equal(t, tt.config.AuthorEmail, cfg.AuthorEmail)
				assert.Equal(t, tt.config.JiraHost, cfg.JiraHost)
				assert.Equal(t, tt.config.JiraToken, cfg.JiraToken)
				assert.Equal(t, tt.config.JiraProject, cfg.JiraProject)
				assert.Equal(t, tt.config.JiraUser, cfg.JiraUser)
			} else {
				// Should return default config
				assert.Equal(t, DefaultGitBranch, cfg.GitBranch)
				assert.Equal(t, DefaultGitRemote, cfg.GitRemote)
				assert.Equal(t, DefaultStoryDir, cfg.StoryDir)
				assert.Equal(t, DefaultPairFile, cfg.PairFile)
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		wantErr  bool
		setup    func(t *testing.T) string
		teardown func(t *testing.T, path string)
	}{
		{
			name: "save_valid_config",
			config: &Config{
				GitRepo:     "test-repo",
				GitBranch:   "main",
				GitRemote:   "origin",
				StoryDir:    "stories",
				PairFile:    "pair.json",
				AuthorName:  "Test User",
				AuthorEmail: "test@example.com",
				JiraHost:    "https://jira.example.com",
				JiraToken:   "token123",
				JiraProject: "TEST",
				JiraUser:    "test@example.com",
			},
			setup: func(t *testing.T) string {
				dir := t.TempDir()

				// Save current directory
				currentDir, err := os.Getwd()
				assert.NoError(t, err)

				// Change to test directory
				err = os.Chdir(dir)
				assert.NoError(t, err)

				// Create .tracer directory
				tracerDir := filepath.Join(dir, ".tracer")
				err = os.MkdirAll(tracerDir, 0755)
				assert.NoError(t, err)

				// Change back to original directory
				err = os.Chdir(currentDir)
				assert.NoError(t, err)

				utils.TestConfigDir = tracerDir
				return dir
			},
			teardown: func(t *testing.T, path string) {
				utils.TestConfigDir = ""
			},
		},
		{
			name: "save_to_invalid_directory",
			config: &Config{
				GitRepo: "test-repo",
			},
			wantErr: true,
			setup: func(t *testing.T) string {
				// Save current directory
				currentDir, err := os.Getwd()
				assert.NoError(t, err)

				// Create a temporary directory that is not a Git repository
				dir := t.TempDir()

				// Change to the non-Git directory
				err = os.Chdir(dir)
				assert.NoError(t, err)

				// Set an invalid path that doesn't exist and can't be created
				utils.TestConfigDir = "/dev/null/invalid"

				return currentDir // Return the original directory to restore later
			},
			teardown: func(t *testing.T, path string) {
				// Restore the original directory
				err := os.Chdir(path)
				assert.NoError(t, err)
				utils.TestConfigDir = ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			defer tt.teardown(t, dir)

			var err error
			var currentDir string
			// Only change directory for valid config test
			if tt.name == "save_valid_config" {
				// Save current directory
				currentDir, err = os.Getwd()
				assert.NoError(t, err)

				// Change to test directory
				err = os.Chdir(dir)
				assert.NoError(t, err)
				defer func() {
					err = os.Chdir(currentDir)
					assert.NoError(t, err)
				}()
			}

			err = SaveConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Verify the file was created and contains the correct data
			var configFile string
			if tt.name == "save_valid_config" {
				configFile = filepath.Join(dir, ".tracer", DefaultConfigFile)
			} else {
				configFile = filepath.Join(utils.TestConfigDir, DefaultConfigFile)
			}

			data, err := os.ReadFile(configFile)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			var loadedConfig Config
			err = yaml.Unmarshal(data, &loadedConfig)
			assert.NoError(t, err)

			assert.Equal(t, tt.config.GitRepo, loadedConfig.GitRepo)
			assert.Equal(t, tt.config.GitBranch, loadedConfig.GitBranch)
			assert.Equal(t, tt.config.GitRemote, loadedConfig.GitRemote)
			assert.Equal(t, tt.config.StoryDir, loadedConfig.StoryDir)
			assert.Equal(t, tt.config.PairFile, loadedConfig.PairFile)
			assert.Equal(t, tt.config.AuthorName, loadedConfig.AuthorName)
			assert.Equal(t, tt.config.AuthorEmail, loadedConfig.AuthorEmail)
			assert.Equal(t, tt.config.JiraHost, loadedConfig.JiraHost)
			assert.Equal(t, tt.config.JiraToken, loadedConfig.JiraToken)
			assert.Equal(t, tt.config.JiraProject, loadedConfig.JiraProject)
			assert.Equal(t, tt.config.JiraUser, loadedConfig.JiraUser)
		})
	}
}

func TestRepositorySpecificConfig(t *testing.T) {
	// Create two separate repositories
	repo1Dir := t.TempDir()
	repo2Dir := t.TempDir()

	// Save current directory
	currentDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err = os.Chdir(currentDir)
		require.NoError(t, err)
	}()

	// Create a mock git client
	mockGit := utils.NewMockGit()
	utils.GitClient = mockGit

	// Configure mock behavior
	mockGit.(*utils.MockGit).GetGitRootFunc = func() (string, error) {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return wd, nil
	}

	// Setup first repository
	err = os.Chdir(repo1Dir)
	require.NoError(t, err)
	err = utils.GitClient.Init()
	require.NoError(t, err)

	// Configure first repository
	cfg1 := &Config{
		GitRepo:    "repo1",
		GitBranch:  "main",
		AuthorName: "user1",
	}
	err = SaveConfig(cfg1)
	require.NoError(t, err)

	// Setup second repository
	err = os.Chdir(repo2Dir)
	require.NoError(t, err)
	err = utils.GitClient.Init()
	require.NoError(t, err)

	// Configure second repository
	cfg2 := &Config{
		GitRepo:    "repo2",
		GitBranch:  "main",
		AuthorName: "user2",
	}
	err = SaveConfig(cfg2)
	require.NoError(t, err)

	// Verify first repository config
	err = os.Chdir(repo1Dir)
	require.NoError(t, err)
	loadedCfg1, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "repo1", loadedCfg1.GitRepo)
	assert.Equal(t, "user1", loadedCfg1.AuthorName)

	// Verify second repository config
	err = os.Chdir(repo2Dir)
	require.NoError(t, err)
	loadedCfg2, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "repo2", loadedCfg2.GitRepo)
	assert.Equal(t, "user2", loadedCfg2.AuthorName)

	// Restore the real git client
	utils.GitClient = utils.NewRealGit()
}
