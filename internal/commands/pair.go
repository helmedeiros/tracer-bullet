package commands

import (
	"fmt"

	"github.com/helmedeiros/tracer-bullet/internal/config"
	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"github.com/spf13/cobra"
)

var PairCmd = &cobra.Command{
	Use:   "pair",
	Short: "Manage pair programming sessions",
	Long: `Manage your pair programming sessions through a natural workflow:

1. Start a Session
   tracer pair start <partner-name>  # Begin pairing with a teammate

2. Check Status
   tracer pair show                 # View current session details

3. End Session
   tracer pair stop                 # End the current session

Each command helps you maintain effective pair programming practices.
Use these commands to:
- Track who you're working with
- Monitor session progress
- Maintain clear development history`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		switch args[0] {
		case "start":
			if len(args) < 2 {
				return fmt.Errorf(`partner name is required. Please provide it as follows:
tracer pair start <partner-name>

Example:
  tracer pair start john.doe`)
			}
			return startPair(cmd, args[1])
		case "stop":
			return stopPair(cmd)
		case "show":
			return showPairStatus(cmd)
		default:
			return fmt.Errorf(`unknown command: %s

Available Commands:
  start <partner-name>  Start a pair programming session
  show                  Show current session status
  stop                  End the current session

Example:
  tracer pair start john.doe`, args[0])
		}
	},
}

func startPair(cmd *cobra.Command, partner string) error {
	if partner == "" {
		return fmt.Errorf(`partner name cannot be empty. Please provide a valid name:
tracer pair start <partner-name>

Example:
  tracer pair start john.doe`)
	}

	// Validate project configuration
	projectName, err := utils.GitClient.GetConfig("current.project")
	if err != nil || projectName == "" {
		return fmt.Errorf(`project not configured. Please follow these steps:
1. Run 'tracer init' to initialize your project
2. Run 'tracer configure project' to set up project settings`)
	}

	// Set git config for pair
	err = utils.GitClient.SetConfig("current.pair", partner)
	if err != nil {
		return fmt.Errorf("failed to set pair: %w", err)
	}

	// Update config file
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cfg.PairName = partner
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Get current user for better context
	currentUser, _ := utils.GitClient.GetConfig(fmt.Sprintf("%s.user", projectName))
	if currentUser == "" {
		currentUser = "unknown user"
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nStarted pair programming session!\n\n")
	fmt.Fprintf(cmd.OutOrStdout(), "Session Details:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  Project: %s\n", projectName)
	fmt.Fprintf(cmd.OutOrStdout(), "  Current User: %s\n", currentUser)
	fmt.Fprintf(cmd.OutOrStdout(), "  Pair Partner: %s\n", partner)
	fmt.Fprintf(cmd.OutOrStdout(), "\nNext steps:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "1. Create a new story with 'tracer story new'\n")
	fmt.Fprintf(cmd.OutOrStdout(), "2. Make changes together\n")
	fmt.Fprintf(cmd.OutOrStdout(), "3. Create commits with 'tracer commit create'\n")
	fmt.Fprintf(cmd.OutOrStdout(), "4. End session with 'tracer pair stop' when done\n")

	return nil
}

func stopPair(cmd *cobra.Command) error {
	// Get current pair info for better output
	pairName, _ := utils.GitClient.GetConfig("current.pair")
	projectName, _ := utils.GitClient.GetConfig("current.project")
	currentUser, _ := utils.GitClient.GetConfig(fmt.Sprintf("%s.user", projectName))

	// Remove git config for pair
	err := utils.GitClient.SetConfig("current.pair", "")
	if err != nil {
		return fmt.Errorf("failed to remove pair: %w", err)
	}

	// Update config file
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cfg.PairName = ""
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nEnded pair programming session!\n\n")
	if pairName != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Session Summary:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  Project: %s\n", projectName)
		fmt.Fprintf(cmd.OutOrStdout(), "  Participants: %s and %s\n", currentUser, pairName)
		fmt.Fprintf(cmd.OutOrStdout(), "\nThanks for pairing! 👥\n")
	}

	return nil
}

func showPairStatus(cmd *cobra.Command) error {
	// Get current pair from git config
	pairName, err := utils.GitClient.GetConfig("current.pair")
	if err != nil || pairName == "" {
		fmt.Fprintf(cmd.OutOrStdout(), "\nNo active pair programming session\n")
		fmt.Fprintf(cmd.OutOrStdout(), "\nTo start a session:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  tracer pair start <partner-name>\n")
		return nil
	}

	// Get project name
	projectName, err := utils.GitClient.GetConfig("current.project")
	if err != nil {
		projectName = "unknown project"
	}

	// Get config for additional context
	cfg, err := config.LoadConfig()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	// Get current user
	currentUser, err := utils.GitClient.GetConfig(fmt.Sprintf("%s.user", projectName))
	if err != nil {
		currentUser = "unknown user"
	}

	// Display detailed pairing information
	fmt.Fprintf(cmd.OutOrStdout(), "\nActive Pair Programming Session:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  Project: %s\n", projectName)
	fmt.Fprintf(cmd.OutOrStdout(), "  Current User: %s\n", currentUser)
	fmt.Fprintf(cmd.OutOrStdout(), "  Pair Partner: %s\n", pairName)

	// If there's a story associated with the pair, show it
	if cfg.JiraHost != "" && cfg.JiraProject != "" {
		storyID, err := utils.GitClient.GetConfig(fmt.Sprintf("%s.current.story", cfg.JiraProject))
		if err == nil && storyID != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  Current Story: %s\n", storyID)
			fmt.Fprintf(cmd.OutOrStdout(), "  Story URL: %s/browse/%s\n", cfg.JiraHost, storyID)
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nAvailable Commands:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  tracer story new     # Create a new story\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  tracer commit create # Create a commit\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  tracer pair stop     # End the session\n")

	return nil
}
