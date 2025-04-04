package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"gopkg.in/yaml.v3"
)

const (
	// Default configuration paths
	DefaultConfigDir  = ".tracer"
	DefaultConfigFile = "config.yaml"

	// Git related constants
	DefaultGitBranch = "main"
	DefaultGitRemote = "origin"

	// Story related constants
	DefaultStoryDir = "stories"
	DefaultStoryExt = ".yaml"

	// Pair programming related constants
	DefaultPairFile = "pair.json"

	// Jira related constants
	DefaultJiraHost      = "" // Must be configured by user
	DefaultJiraProject   = "" // Must be configured by user
	DefaultJiraIssueType = "Story"
)

// Config represents the application configuration
type Config struct {
	GitRepo     string `yaml:"git_repo"`
	GitBranch   string `yaml:"git_branch"`
	GitRemote   string `yaml:"git_remote"`
	StoryDir    string `yaml:"story_dir"`
	PairFile    string `yaml:"pair_file"`
	AuthorName  string `yaml:"author_name"`
	AuthorEmail string `yaml:"author_email"`
	PairName    string `yaml:"pair_name"`
	JiraHost    string `yaml:"jira_host"`
	JiraToken   string `yaml:"jira_token"`
	JiraProject string `yaml:"jira_project"`
	JiraUser    string `yaml:"jira_user"`
}

// DefaultConfig returns a new Config with default values
func DefaultConfig() *Config {
	return &Config{
		GitBranch: DefaultGitBranch,
		GitRemote: DefaultGitRemote,
		StoryDir:  DefaultStoryDir,
		PairFile:  DefaultPairFile,
	}
}

// setDefaultValues sets default values for empty fields in the config
func setDefaultValues(cfg *Config) {
	if cfg.GitBranch == "" {
		cfg.GitBranch = DefaultGitBranch
	}
	if cfg.GitRemote == "" {
		cfg.GitRemote = DefaultGitRemote
	}
	if cfg.StoryDir == "" {
		cfg.StoryDir = DefaultStoryDir
	}
	if cfg.PairFile == "" {
		cfg.PairFile = DefaultPairFile
	}
}

// loadConfigFromFile loads and unmarshals config from the given file path
func loadConfigFromFile(configFile string) (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	setDefaultValues(&cfg)
	return &cfg, nil
}

// LoadConfig loads the configuration from the config file
func LoadConfig() (*Config, error) {
	// Try to get repository-specific config first
	repoConfigDir, err := utils.GetRepoConfigDir()
	if err == nil {
		repoConfigFile := filepath.Join(repoConfigDir, DefaultConfigFile)
		cfg, err := loadConfigFromFile(repoConfigFile)
		if err == nil {
			return cfg, nil
		}
	}

	// If no repository-specific config exists, try global config
	globalConfigDir, err := utils.GetConfigDir()
	if err != nil {
		return nil, err
	}

	globalConfigFile := filepath.Join(globalConfigDir, DefaultConfigFile)
	cfg, err := loadConfigFromFile(globalConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if no config exists
			cfg := DefaultConfig()
			setDefaultValues(cfg)
			return cfg, nil
		}
		return nil, err
	}

	return cfg, nil
}

// SaveConfig saves the configuration to the config file
func SaveConfig(cfg *Config) error {
	// Try to save to repository-specific config first
	repoConfigDir, err := utils.GetRepoConfigDir()
	if err == nil {
		// Check if the directory is writable
		if err := checkDirWritable(repoConfigDir); err != nil {
			return err
		}

		// Create config directory if it doesn't exist
		if err := os.MkdirAll(repoConfigDir, utils.DefaultDirPerm); err != nil {
			return err
		}

		repoConfigFile := filepath.Join(repoConfigDir, DefaultConfigFile)
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}

		return os.WriteFile(repoConfigFile, data, utils.DefaultFilePerm)
	}

	// If no repository is found, save to global config
	globalConfigDir, err := utils.GetConfigDir()
	if err != nil {
		return err
	}

	// Check if the directory is writable
	if err := checkDirWritable(globalConfigDir); err != nil {
		return err
	}

	// Create config directory if it doesn't exist and it's not a test directory
	if utils.TestConfigDir == "" {
		if err := os.MkdirAll(globalConfigDir, utils.DefaultDirPerm); err != nil {
			return err
		}
	}

	globalConfigFile := filepath.Join(globalConfigDir, DefaultConfigFile)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(globalConfigFile, data, utils.DefaultFilePerm)
}

// checkDirWritable checks if a directory is writable by attempting to create a temporary file
func checkDirWritable(dir string) error {
	// First check if the path exists and is a directory
	if info, err := os.Stat(dir); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("path exists but is not a directory: %s", dir)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check directory: %w", err)
	}

	// If the directory doesn't exist and it's not a test directory, try to create it
	if utils.TestConfigDir == "" {
		if err := os.MkdirAll(dir, utils.DefaultDirPerm); err != nil {
			return fmt.Errorf("directory does not exist and cannot be created: %w", err)
		}
	}

	// Try to create a temporary file to check if the directory is writable
	tmpFile := filepath.Join(dir, ".tracer_test_"+utils.GenerateID())
	if err := os.WriteFile(tmpFile, []byte("test"), utils.DefaultFilePerm); err != nil {
		return fmt.Errorf("directory is not writable: %w", err)
	}

	// Clean up the temporary file
	if err := os.Remove(tmpFile); err != nil {
		return fmt.Errorf("failed to remove temporary file: %w", err)
	}

	return nil
}
