package config

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
