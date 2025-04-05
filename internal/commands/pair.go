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
	Long:  `Start, stop, or check the status of pair programming sessions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		switch args[0] {
		case "start":
			if len(args) < 2 {
				return fmt.Errorf("partner name is required")
			}
			return startPair(cmd, args[1])
		case "stop":
			return stopPair(cmd)
		case "show":
			return showPairStatus(cmd)
		default:
			return fmt.Errorf("unknown command: %s", args[0])
		}
	},
}

func startPair(cmd *cobra.Command, partner string) error {
	if partner == "" {
		return fmt.Errorf("partner name cannot be empty")
	}

	// Set git config for pair
	err := utils.GitClient.SetConfig("current.pair", partner)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Started pair programming session with %s\n", partner)
	return nil
}

func stopPair(cmd *cobra.Command) error {
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

	fmt.Fprintf(cmd.OutOrStdout(), "Stopped pair programming session\n")
	return nil
}

func showPairStatus(cmd *cobra.Command) error {
	// Get current pair from git config
	pairName, err := utils.GitClient.GetConfig("current.pair")
	if err != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "No active pair programming session\n")
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
	fmt.Fprintf(cmd.OutOrStdout(), "\nPair Programming Session:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  Project: %s\n", projectName)
	fmt.Fprintf(cmd.OutOrStdout(), "  Current User: %s\n", currentUser)
	fmt.Fprintf(cmd.OutOrStdout(), "  Pair Partner: %s\n", pairName)

	// If there's a story associated with the pair, show it
	if cfg.JiraHost != "" && cfg.JiraProject != "" {
		storyID, err := utils.GitClient.GetConfig(fmt.Sprintf("%s.current.story", cfg.JiraProject))
		if err == nil && storyID != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  Current Story: %s\n", storyID)
		}
	}
	fmt.Fprintf(cmd.OutOrStdout(), "\n")

	return nil
}
