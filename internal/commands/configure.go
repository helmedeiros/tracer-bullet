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

		// Check if project flag is set but empty
		if cmd.Flags().Changed("project") {
			if err := configureProject(projectFlag); err != nil {
				return fmt.Errorf("failed to configure project: %w", err)
			}
			fmt.Printf("Successfully configured project: %s\n", projectFlag)
		}

		if cmd.Flags().Changed("user") {
			if err := configureUser(userFlag); err != nil {
				return fmt.Errorf("failed to configure user: %w", err)
			}
			fmt.Printf("Successfully configured user: %s\n", userFlag)
		}

		if !autocompleteFlag && !cmd.Flags().Changed("project") && !cmd.Flags().Changed("user") {
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

var configureCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove configurations",
	Long:  `Remove tracer configurations. Use subcommands to clean specific configurations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var configureCleanGitCmd = &cobra.Command{
	Use:   "git",
	Short: "Remove git configurations",
	Long:  `Remove all git-related configurations, including project and user settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get current project name to clear project-specific configs
		projectName, err := utils.GitClient.GetConfig("current.project")
		if err == nil && projectName != "" {
			// Clear project-specific user config
			if err := utils.GitClient.SetConfig(fmt.Sprintf("%s.user", projectName), ""); err != nil {
				return fmt.Errorf("failed to clear project user config: %w", err)
			}
		}

		// Clear git config
		if err := utils.GitClient.SetConfig("current.project", ""); err != nil {
			return fmt.Errorf("failed to clear project config: %w", err)
		}
		if err := utils.GitClient.SetConfig("current.pair", ""); err != nil {
			return fmt.Errorf("failed to clear pair config: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Git configurations have been removed\n")
		return nil
	},
}

var configureCleanStoriesCmd = &cobra.Command{
	Use:   "stories",
	Short: "Remove story configurations",
	Long:  `Remove all story-related configurations and data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get repository-specific config directory first
		repoConfigDir, err := utils.GetRepoConfigDir()
		if err == nil {
			// Remove stories directory
			storiesDir := filepath.Join(repoConfigDir, "stories")
			if err := os.RemoveAll(storiesDir); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove stories directory: %w", err)
			}
		}

		// Get global config directory
		globalConfigDir, err := utils.GetConfigDir()
		if err != nil {
			return fmt.Errorf("failed to get config directory: %w", err)
		}

		// Remove stories directory
		storiesDir := filepath.Join(globalConfigDir, "stories")
		if err := os.RemoveAll(storiesDir); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove stories directory: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Story configurations have been removed\n")
		return nil
	},
}

var configureCleanJiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Remove Jira configurations",
	Long:  `Remove all Jira-related configurations, including host, project, and authentication settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load current config
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Clear Jira settings
		cfg.JiraHost = ""
		cfg.JiraToken = ""
		cfg.JiraProject = ""
		cfg.JiraUser = ""

		// Save updated config
		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Jira configurations have been removed\n")
		return nil
	},
}

var configureCleanAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Remove all configurations",
	Long:  `Remove all tracer configurations, including project, user, Jira settings, and stories.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get current project name to clear project-specific configs
		projectName, err := utils.GitClient.GetConfig("current.project")
		if err == nil && projectName != "" {
			// Clear project-specific user config
			if err := utils.GitClient.SetConfig(fmt.Sprintf("%s.user", projectName), ""); err != nil {
				return fmt.Errorf("failed to clear project user config: %w", err)
			}
		}

		// Get repository-specific config directory first
		repoConfigDir, err := utils.GetRepoConfigDir()
		if err == nil {
			// Remove repository-specific config file
			repoConfigFile := filepath.Join(repoConfigDir, config.DefaultConfigFile)
			if err := os.Remove(repoConfigFile); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove repository config: %w", err)
			}
			// Remove stories directory
			storiesDir := filepath.Join(repoConfigDir, "stories")
			if err := os.RemoveAll(storiesDir); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove stories directory: %w", err)
			}
		}

		// Get global config directory
		globalConfigDir, err := utils.GetConfigDir()
		if err != nil {
			return fmt.Errorf("failed to get config directory: %w", err)
		}

		// Remove global config file
		globalConfigFile := filepath.Join(globalConfigDir, config.DefaultConfigFile)
		if err := os.Remove(globalConfigFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove global config: %w", err)
		}
		// Remove stories directory
		storiesDir := filepath.Join(globalConfigDir, "stories")
		if err := os.RemoveAll(storiesDir); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove stories directory: %w", err)
		}

		// Clear git config
		if err := utils.GitClient.SetConfig("current.project", ""); err != nil {
			return fmt.Errorf("failed to clear project config: %w", err)
		}
		if err := utils.GitClient.SetConfig("current.pair", ""); err != nil {
			return fmt.Errorf("failed to clear pair config: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "All configurations have been removed\n")
		return nil
	},
}

func init() {
	ConfigureCmd.Flags().StringVarP(&projectFlag, "project", "p", "", "Set the project name")
	ConfigureCmd.Flags().StringVarP(&userFlag, "user", "u", "", "Set the user name")
	ConfigureCmd.Flags().BoolVarP(&autocompleteFlag, "autocomplete", "a", false, "Configure zsh autocomplete")

	// Add subcommands
	ConfigureCmd.AddCommand(configureShowCmd)
	ConfigureCmd.AddCommand(configureCleanCmd)

	// Add subcommands to configure clean
	configureCleanCmd.AddCommand(configureCleanGitCmd)
	configureCleanCmd.AddCommand(configureCleanStoriesCmd)
	configureCleanCmd.AddCommand(configureCleanJiraCmd)
	configureCleanCmd.AddCommand(configureCleanAllCmd)
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
	err = utils.GitClient.SetConfig("current.project", projectName)
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
	projectName, err := utils.GitClient.GetConfig("current.project")
	if err != nil {
		return fmt.Errorf("project not configured. Please run 'tracer configure --project' first")
	}

	// If no project is configured, return error
	if projectName == "" {
		return fmt.Errorf("project not configured. Please run 'tracer configure --project' first")
	}

	// Set git config for user
	err = utils.GitClient.SetConfig(fmt.Sprintf("%s.user", projectName), username)
	if err != nil {
		return fmt.Errorf("failed to set git config for user: %w", err)
	}

	// Try to get repository-specific config directory
	configDir, err := utils.GetRepoConfigDir()
	if err != nil {
		// If we can't get the repo config dir, use the global config dir
		configDir, err = utils.GetConfigDir()
		if err != nil {
			return fmt.Errorf("failed to get config directory: %w", err)
		}
	}

	// Create or update config file
	cfg := &config.Config{
		GitRepo:    projectName,
		GitBranch:  config.DefaultGitBranch,
		GitRemote:  config.DefaultGitRemote,
		AuthorName: username,
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

// checkZshrcConfig checks if TRACER_HOME and autocomplete are already configured in .zshrc
func checkZshrcConfig(lines []string, exportLine string) (bool, bool) {
	tracerHomeConfigured := false
	autocompleteConfigured := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == strings.TrimSpace(exportLine) {
			tracerHomeConfigured = true
		}
		if trimmedLine == "fpath=($TRACER_HOME/completion/zsh $fpath)" {
			autocompleteConfigured = true
		}
	}

	return tracerHomeConfigured, autocompleteConfigured
}

// updateZshrcContent updates .zshrc content with TRACER_HOME and autocomplete configuration
func updateZshrcContent(lines []string, exportLine string, tracerHomeConfigured, autocompleteConfigured bool) []string {
	updatedLines := make([]string, 0, len(lines)+3)
	updatedLines = append(updatedLines, lines...)

	// Add a newline if the file doesn't end with one
	if len(updatedLines) > 0 && updatedLines[len(updatedLines)-1] != "" {
		updatedLines = append(updatedLines, "")
	}

	// Add TRACER_HOME if not configured
	if !tracerHomeConfigured {
		updatedLines = append(updatedLines, exportLine)
	}

	// Add autocomplete configuration if not configured
	if !autocompleteConfigured {
		updatedLines = append(updatedLines,
			"fpath=($TRACER_HOME/completion/zsh $fpath)",
			"autoload -U compinit",
			"compinit",
		)
	}

	return updatedLines
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

	// Get absolute path for completion script
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Set TRACER_HOME environment variable
	if err := os.Setenv("TRACER_HOME", cwd); err != nil {
		return fmt.Errorf("failed to set TRACER_HOME: %w", err)
	}

	zshrcPath := filepath.Join(homeDir, ".zshrc")
	zshrcContent, err := os.ReadFile(zshrcPath)
	if err != nil {
		return fmt.Errorf("failed to read .zshrc: %w", err)
	}

	lines := strings.Split(string(zshrcContent), "\n")
	exportLine := "export TRACER_HOME=" + cwd

	tracerHomeConfigured, autocompleteConfigured := checkZshrcConfig(lines, exportLine)

	if tracerHomeConfigured && autocompleteConfigured {
		fmt.Println("Autocomplete already configured")
		return nil
	}

	updatedLines := updateZshrcContent(lines, exportLine, tracerHomeConfigured, autocompleteConfigured)

	// Write the updated content back to the file
	if err := os.WriteFile(zshrcPath, []byte(strings.Join(updatedLines, "\n")), 0600); err != nil {
		return fmt.Errorf("failed to write .zshrc: %w", err)
	}

	return nil
}
