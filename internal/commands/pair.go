package commands

import (
	"bytes"
	"fmt"

	"github.com/helmedeiros/tracer-bullet/internal/config"
	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"github.com/spf13/cobra"
)

var outputBuffer = &bytes.Buffer{}

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
			return startPair(args[1])
		case "stop":
			return stopPair()
		case "status":
			return showPairStatus()
		default:
			return fmt.Errorf("unknown command: %s", args[0])
		}
	},
}

func startPair(partner string) error {
	if partner == "" {
		return fmt.Errorf("partner name cannot be empty")
	}

	// Set git config for pair
	_, err := utils.RunCommand("git", "config", "--local", "current.pair", partner)
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

	fmt.Fprintf(outputBuffer, "Started pair programming session with %s\n", partner)
	return nil
}

func stopPair() error {
	// Remove git config for pair
	_, err := utils.RunCommand("git", "config", "--local", "--unset", "current.pair")
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

	fmt.Fprintf(outputBuffer, "Stopped pair programming session\n")
	return nil
}

func showPairStatus() error {
	// Get current pair from git config
	pairName, err := utils.RunCommand("git", "config", "--local", "current.pair")
	if err != nil {
		fmt.Fprintf(outputBuffer, "No active pair programming session\n")
		return nil
	}

	fmt.Fprintf(outputBuffer, "Current pair: %s\n", pairName)
	return nil
}

// GetOutput returns the current output buffer content and clears it
func GetOutput() string {
	output := outputBuffer.String()
	outputBuffer.Reset()
	return output
}
