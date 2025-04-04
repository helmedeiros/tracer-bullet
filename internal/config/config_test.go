package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"github.com/stretchr/testify/assert"
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
				utils.TestConfigDir = dir
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
				utils.TestConfigDir = dir
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

			if tt.config != nil {
				// Write config file
				configFile := filepath.Join(dir, DefaultConfigFile)
				data, err := yaml.Marshal(tt.config)
				assert.NoError(t, err)
				err = os.WriteFile(configFile, data, 0644)
				assert.NoError(t, err)
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
			name: "save valid config",
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
				utils.TestConfigDir = dir
				return dir
			},
			teardown: func(t *testing.T, path string) {
				utils.TestConfigDir = ""
			},
		},
		{
			name: "save to invalid directory",
			config: &Config{
				GitRepo: "test-repo",
			},
			wantErr: true,
			setup: func(t *testing.T) string {
				utils.TestConfigDir = "/invalid/path"
				return "/invalid/path"
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

			err := SaveConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Verify the file was created and contains the correct data
			configFile := filepath.Join(dir, DefaultConfigFile)
			data, err := os.ReadFile(configFile)
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
