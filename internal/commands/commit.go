package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/helmedeiros/tracer-bullet/internal/config"
	"github.com/helmedeiros/tracer-bullet/internal/story"
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
		includeJira, _ := cmd.Flags().GetBool("jira")

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

		// Add Jira story URL if requested
		if includeJira {
			cfg, err := config.LoadConfig()
			if err == nil && cfg.JiraHost != "" && cfg.JiraProject != "" {
				// Get current story from git config
				storyID, err := utils.RunCommand("git", "config", "--local", fmt.Sprintf("%s.current.story", cfg.JiraProject))
				if err == nil && storyID != "" {
					commitMsg.WriteString("\n\nJira: https://")
					commitMsg.WriteString(cfg.JiraHost)
					commitMsg.WriteString("/browse/")
					commitMsg.WriteString(cfg.JiraProject)
					commitMsg.WriteString("-")
					commitMsg.WriteString(storyID)
				}
			}
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

		// Get the commit hash
		commitHash, err := utils.RunCommand("git", "rev-parse", "HEAD")
		if err != nil {
			return fmt.Errorf("failed to get commit hash: %w", err)
		}

		// Get the author
		author, err := utils.RunCommand("git", "config", "user.name")
		if err != nil {
			return fmt.Errorf("failed to get author: %w", err)
		}

		// If we have a current story, associate this commit with it
		cfg, err := config.LoadConfig()
		if err == nil && cfg.JiraHost != "" && cfg.JiraProject != "" {
			storyID, err := utils.RunCommand("git", "config", "--local", fmt.Sprintf("%s.current.story", cfg.JiraProject))
			if err == nil && storyID != "" {
				// Load the story
				s, err := story.LoadStory(storyID)
				if err == nil {
					// Add the commit to the story
					s.AddCommit(commitHash, commitMsg.String(), author, time.Now())

					// Get changed files
					files, err := utils.RunCommand("git", "diff", "--name-status", "HEAD~1", "HEAD")
					if err == nil {
						for _, line := range strings.Split(files, "\n") {
							if line == "" {
								continue
							}
							parts := strings.Fields(line)
							if len(parts) >= 2 {
								status := parts[0]
								path := parts[1]
								s.AddFile(path, status)
							}
						}
					}

					// Save the updated story
					if err := s.Save(); err != nil {
						return fmt.Errorf("failed to update story: %w", err)
					}
				}
			}
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
	commitCreateCmd.Flags().Bool("jira", false, "Include Jira story URL in commit body")

	// Handle required flags
	requiredFlags := []string{"type", "message"}
	for _, flag := range requiredFlags {
		if err := commitCreateCmd.MarkFlagRequired(flag); err != nil {
			// Since this is during initialization, panic is appropriate
			panic(fmt.Sprintf("failed to mark %s flag as required: %v", flag, err))
		}
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
