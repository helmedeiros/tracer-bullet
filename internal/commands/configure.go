package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/helmedeiros/tracer-bullet/internal/config"
	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	projectFlag      string
	userFlag         string
	autocompleteFlag bool
)

var ConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure tracer settings",
	Long:  `Configure various settings for the tracer tool, including git repository settings, story tracking preferences, and shell autocomplete.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if autocompleteFlag {
			if err := configureAutocomplete(); err != nil {
				return fmt.Errorf("failed to configure autocomplete: %w", err)
			}
			fmt.Println("Successfully configured zsh autocomplete")
		}

		if projectFlag != "" {
			if err := configureProject(projectFlag); err != nil {
				return fmt.Errorf("failed to configure project: %w", err)
			}
			fmt.Printf("Successfully configured project: %s\n", projectFlag)
		}

		if userFlag != "" {
			if err := configureUser(userFlag); err != nil {
				return fmt.Errorf("failed to configure user: %w", err)
			}
			fmt.Printf("Successfully configured user: %s\n", userFlag)
		}

		if !autocompleteFlag && projectFlag == "" && userFlag == "" {
			return cmd.Help()
		}

		return nil
	},
}

var configureShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current configuration settings for the tracer tool.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load current config
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Print configuration
		fmt.Fprintf(cmd.OutOrStdout(), "Current Configuration:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "Project: %s\n", cfg.GitRepo)
		fmt.Fprintf(cmd.OutOrStdout(), "User: %s\n", cfg.AuthorName)
		fmt.Fprintf(cmd.OutOrStdout(), "Jira:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  Host: %s\n", cfg.JiraHost)
		fmt.Fprintf(cmd.OutOrStdout(), "  Project: %s\n", cfg.JiraProject)
		fmt.Fprintf(cmd.OutOrStdout(), "  User: %s\n", cfg.JiraUser)
		if cfg.JiraToken != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  Token: [CONFIGURED]\n")
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "  Token: [NOT CONFIGURED]\n")
		}

		return nil
	},
}

func init() {
	ConfigureCmd.Flags().StringVarP(&projectFlag, "project", "p", "", "Set the project name")
	ConfigureCmd.Flags().StringVarP(&userFlag, "user", "u", "", "Set the user name")
	ConfigureCmd.Flags().BoolVarP(&autocompleteFlag, "autocomplete", "a", false, "Configure zsh autocomplete")

	// Add show subcommand
	ConfigureCmd.AddCommand(configureShowCmd)
}

func configureProject(projectName string) error {
	if projectName == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	// Get repository-specific config directory first
	configDir, err := utils.GetRepoConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get repo config directory: %w", err)
	}

	// Set git config
	_, err = utils.RunCommand("git", "config", "--local", "current.project", projectName)
	if err != nil {
		return fmt.Errorf("failed to set git config: %w", err)
	}

	// Create or update config file
	cfg := &config.Config{
		GitRepo:   projectName,
		GitBranch: config.DefaultGitBranch,
		GitRemote: config.DefaultGitRemote,
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configFile := filepath.Join(configDir, config.DefaultConfigFile)
	if err := os.WriteFile(configFile, data, utils.DefaultFilePerm); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func configureUser(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Get current project name
	projectName, err := utils.RunCommand("git", "config", "--local", "current.project")
	if err != nil {
		return fmt.Errorf("project not configured. Please run 'tracer configure project' first")
	}

	// Set git config for user
	_, err = utils.RunCommand("git", "config", "--local", fmt.Sprintf("%s.user", projectName), username)
	if err != nil {
		return fmt.Errorf("failed to set git config for user: %w", err)
	}

	// Get repository-specific config directory
	configDir, err := utils.GetRepoConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get repo config directory: %w", err)
	}

	// Update config file
	configFile := filepath.Join(configDir, config.DefaultConfigFile)
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("project not configured. Please run 'tracer configure project' first")
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	cfg.AuthorName = username

	data, err = yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, data, utils.DefaultFilePerm); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func configureAutocomplete() error {
	// Generate completion files
	if err := GenerateZshCompletion(); err != nil {
		return fmt.Errorf("failed to generate completion files: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	zshrcPath := filepath.Join(homeDir, ".zshrc")
	zshrcContent, err := os.ReadFile(zshrcPath)
	if err != nil {
		return fmt.Errorf("failed to read .zshrc: %w", err)
	}

	// Check if autocomplete is already configured
	configLine := fmt.Sprintf("fpath=(%s/completion/zsh $fpath)", os.Getenv("BASEDIR"))
	if strings.Contains(string(zshrcContent), configLine) {
		fmt.Println("Autocomplete already configured")
		return nil
	}

	// Append autocomplete configuration
	autocompleteConfig := fmt.Sprintf(`
# Tracer autocomplete configuration
fpath=(%s/completion/zsh $fpath)
autoload -U compinit
compinit
`, os.Getenv("BASEDIR"))

	file, err := os.OpenFile(zshrcPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open .zshrc: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(autocompleteConfig); err != nil {
		return fmt.Errorf("failed to write to .zshrc: %w", err)
	}

	return nil
}
