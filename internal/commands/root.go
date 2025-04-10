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
	Long: `Tracer Bullet helps developers manage their workflow through a natural progression:

1. Setup: Initialize and configure your environment
   tracer init           # Initialize a new project
   tracer configure     # Set up your environment

2. Work: Create and track stories, manage commits
   tracer story        # Manage development stories
   tracer commit       # Create and manage commits

3. Collaborate: Handle pair programming sessions
   tracer pair         # Manage pair programming

4. Integrate: Connect with external tools
   tracer jira         # Jira integration

Each command follows a natural workflow, making it easy to:
- Start new projects
- Track your work
- Collaborate with others
- Manage your development process`,
}

func init() {
	// Add all commands to root
	RootCmd.AddCommand(InitCmd)
	RootCmd.AddCommand(ConfigureCmd)
	RootCmd.AddCommand(StoryCmd)
	RootCmd.AddCommand(CommitCmd)
	RootCmd.AddCommand(PairCmd)
	RootCmd.AddCommand(JiraCmd)
}

// Execute runs the root command
func Execute() error {
	return RootCmd.Execute()
}

// GenerateZshCompletion generates zsh completion files
func GenerateZshCompletion() error {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Get the completion directory
	completionDir := filepath.Join(cwd, "completion", "zsh")
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
