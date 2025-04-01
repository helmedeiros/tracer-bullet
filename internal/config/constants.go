package config

import (
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
	DefaultStoryExt = ".md"

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

// LoadConfig loads the configuration from the config file
func LoadConfig() (*Config, error) {
	configDir, err := utils.GetConfigDir()
	if err != nil {
		return nil, err
	}

	configFile := filepath.Join(configDir, DefaultConfigFile)
	data, err := os.ReadFile(configFile)
	if err != nil {
		return &Config{}, nil // Return default config if file doesn't exist
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// SaveConfig saves the configuration to the config file
func SaveConfig(cfg *Config) error {
	configDir, err := utils.GetConfigDir()
	if err != nil {
		return err
	}

	configFile := filepath.Join(configDir, DefaultConfigFile)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, utils.DefaultFilePerm)
}
