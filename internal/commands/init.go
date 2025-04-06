package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/helmedeiros/tracer-bullet/internal/config"
	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new tracer configuration in the current directory",
	Long: `Initialize a new tracer configuration in the current directory.
This will create a .tracer directory and set up the necessary configuration files.`,
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	// Create .tracer directory
	tracerDir := ".tracer"
	if err := os.MkdirAll(tracerDir, 0755); err != nil {
		return fmt.Errorf("failed to create .tracer directory: %w", err)
	}

	// Create default config using the config package
	cfg := config.DefaultConfig()

	// Try to get git repository name
	if gitRoot, err := utils.GitClient.GetGitRoot(); err == nil {
		cfg.GitRepo = filepath.Base(gitRoot)
	}

	// Try to get git user info
	if userName, err := utils.GitClient.GetConfig("user.name"); err == nil {
		cfg.AuthorName = userName
	}
	if userEmail, err := utils.GitClient.GetConfig("user.email"); err == nil {
		cfg.AuthorEmail = userEmail
	}

	// Save the configuration to the repository-specific directory
	configPath := filepath.Join(tracerDir, config.DefaultConfigFile)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(configPath, data, utils.DefaultFilePerm); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Println("Initialized tracer configuration in", tracerDir)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Run 'tracer configure --project' to set up project-specific settings")
	fmt.Println("2. Run 'tracer configure --user' to set up your user information")
	fmt.Println("3. Run 'tracer configure' to set up your Jira credentials")

	return nil
}
