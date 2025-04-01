package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"github.com/spf13/cobra"
)

var CommitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Create a conventional commit",
	Long: `Create a git commit following conventional commit format.
Supports common types like feat, fix, docs, style, refactor, test, and chore.
Automatically includes scope based on current story if available.`,
}

var commitCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new commit",
	Long: `Create a new git commit with proper conventional commit format.
Example: tracer commit create --type feat --scope auth --message "add login functionality"
Will create: feat(auth): add login functionality`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flag values
		commitType, _ := cmd.Flags().GetString("type")
		scope, _ := cmd.Flags().GetString("scope")
		message, _ := cmd.Flags().GetString("message")
		body, _ := cmd.Flags().GetString("body")
		breaking, _ := cmd.Flags().GetBool("breaking")

		// Validate commit type
		if !isValidCommitType(commitType) {
			return fmt.Errorf("invalid commit type: %s. Must be one of: feat, fix, docs, style, refactor, test, chore", commitType)
		}

		// Validate message
		if message == "" {
			return fmt.Errorf("commit message cannot be empty")
		}

		// Build commit message
		var commitMsg strings.Builder

		// First line: type(scope): message
		commitMsg.WriteString(commitType)
		if scope != "" {
			commitMsg.WriteString(fmt.Sprintf("(%s)", scope))
		}
		if breaking {
			commitMsg.WriteString("!")
		}
		commitMsg.WriteString(fmt.Sprintf(": %s", message))

		// Add body if provided
		if body != "" {
			commitMsg.WriteString("\n\n")
			commitMsg.WriteString(body)
		}

		// Add breaking change footer if needed
		if breaking {
			commitMsg.WriteString("\n\nBREAKING CHANGE: ")
			if !strings.Contains(strings.ToLower(body), "breaking change") {
				commitMsg.WriteString(message)
			}
		}

		// Create temporary file for commit message
		configDir, err := utils.GetConfigDir()
		if err != nil {
			return fmt.Errorf("failed to get config directory: %w", err)
		}

		tmpFile := filepath.Join(configDir, "COMMIT_MSG")
		if err := os.WriteFile(tmpFile, []byte(commitMsg.String()), 0600); err != nil {
			return fmt.Errorf("failed to write commit message: %w", err)
		}
		defer os.Remove(tmpFile)

		// Run git commit
		_, err = utils.RunCommand("git", "commit", "-F", tmpFile)
		if err != nil {
			return fmt.Errorf("failed to create commit: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Created commit: %s\n", commitMsg.String())
		return nil
	},
}

func init() {
	CommitCmd.AddCommand(commitCreateCmd)

	commitCreateCmd.Flags().String("type", "", "Commit type (feat, fix, docs, style, refactor, test, chore)")
	commitCreateCmd.Flags().String("message", "", "Commit message")
	commitCreateCmd.Flags().String("scope", "", "Commit scope (optional)")
	commitCreateCmd.Flags().String("body", "", "Commit body (optional)")
	commitCreateCmd.Flags().Bool("breaking", false, "Mark as breaking change")

	if err := commitCreateCmd.MarkFlagRequired("type"); err != nil {
		panic(fmt.Sprintf("failed to mark type flag as required: %v", err))
	}
	if err := commitCreateCmd.MarkFlagRequired("message"); err != nil {
		panic(fmt.Sprintf("failed to mark message flag as required: %v", err))
	}
}

func isValidCommitType(commitType string) bool {
	validTypes := map[string]bool{
		"feat":     true,
		"fix":      true,
		"docs":     true,
		"style":    true,
		"refactor": true,
		"test":     true,
		"chore":    true,
	}
	return validTypes[commitType]
}
