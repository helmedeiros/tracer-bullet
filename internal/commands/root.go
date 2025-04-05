package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "tracer",
	Short: "Tracer Bullet - A developer productivity tool",
	Long: `Tracer Bullet helps developers manage their workflow by providing tools for:
- Story tracking and management
- Pair programming session management
- Jira integration and synchronization
- Git workflow automation
- Conventional commit support`,
}

func init() {
	// Add all commands to root
	RootCmd.AddCommand(ConfigureCmd)
	RootCmd.AddCommand(PairCmd)
	RootCmd.AddCommand(StoryCmd)
	RootCmd.AddCommand(JiraCmd)
	RootCmd.AddCommand(CommitCmd)
}

// Execute runs the root command
func Execute() error {
	return RootCmd.Execute()
}

// GenerateZshCompletion generates zsh completion files
func GenerateZshCompletion() error {
	// Get the completion directory
	completionDir := filepath.Join(os.Getenv("BASEDIR"), "completion", "zsh")
	if err := os.MkdirAll(completionDir, 0755); err != nil {
		return fmt.Errorf("failed to create completion directory: %w", err)
	}

	// Generate completion file
	completionFile := filepath.Join(completionDir, "_tracer")
	file, err := os.Create(completionFile)
	if err != nil {
		return fmt.Errorf("failed to create completion file: %w", err)
	}
	defer file.Close()

	// Generate completion script
	if err := RootCmd.GenZshCompletion(file); err != nil {
		return fmt.Errorf("failed to generate zsh completion: %w", err)
	}

	return nil
}
