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

// buildCommitMessage builds the commit message from the provided parameters
func buildCommitMessage(commitType, scope, message, body string, breaking bool) string {
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

	return commitMsg.String()
}

// addJiraUrl adds the Jira story URL to the commit message if requested
func addJiraUrl(commitMsg string) (string, error) {
	cfg, err := config.LoadConfig()
	if err != nil || cfg.JiraHost == "" || cfg.JiraProject == "" {
		return commitMsg, nil
	}

	// Get current story from git config
	storyID, err := utils.GitClient.GetConfig(fmt.Sprintf("%s.current.story", cfg.JiraProject))
	if err != nil || storyID == "" {
		return commitMsg, nil
	}

	// Add Jira URL to commit message
	jiraUrl := fmt.Sprintf("\n\nJira: https://%s/browse/%s-%s", cfg.JiraHost, cfg.JiraProject, storyID)
	return commitMsg + jiraUrl, nil
}

// addBreakingChange adds the breaking change footer if needed
func addBreakingChange(commitMsg, message, body string, breaking bool) string {
	if !breaking {
		return commitMsg
	}

	commitMsg += "\n\nBREAKING CHANGE: "
	if !strings.Contains(strings.ToLower(body), "breaking change") {
		commitMsg += message
	}
	return commitMsg
}

// validateCommitType checks if the commit type is valid
func validateCommitType(commitType string) error {
	validTypes := map[string]bool{
		"feat":     true,
		"fix":      true,
		"docs":     true,
		"style":    true,
		"refactor": true,
		"test":     true,
		"chore":    true,
	}

	if !validTypes[commitType] {
		return fmt.Errorf("invalid commit type: %s. Must be one of: feat, fix, docs, style, refactor, test, chore", commitType)
	}
	return nil
}

var commitCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new commit",
	Long:  `Create a new commit with a conventional commit message.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flag values
		commitType, _ := cmd.Flags().GetString("type")
		scope, _ := cmd.Flags().GetString("scope")
		message, _ := cmd.Flags().GetString("message")
		body, _ := cmd.Flags().GetString("body")
		breaking, _ := cmd.Flags().GetBool("breaking")
		includeJira, _ := cmd.Flags().GetBool("jira")

		// Validate commit type
		if err := validateCommitType(commitType); err != nil {
			return err
		}

		// Validate message
		if message == "" {
			return fmt.Errorf("commit message cannot be empty")
		}

		// Build commit message
		commitMsg := buildCommitMessage(commitType, scope, message, body, breaking)

		// Add Jira story URL if requested
		if includeJira {
			var err error
			commitMsg, err = addJiraUrl(commitMsg)
			if err != nil {
				return err
			}
		}

		// Add breaking change footer if needed
		commitMsg = addBreakingChange(commitMsg, message, body, breaking)

		// Create temporary file for commit message
		configDir, err := utils.GetConfigDir()
		if err != nil {
			return fmt.Errorf("failed to get config directory: %w", err)
		}

		tmpFile := filepath.Join(configDir, "COMMIT_MSG")
		if err := os.WriteFile(tmpFile, []byte(commitMsg), 0600); err != nil {
			return fmt.Errorf("failed to write commit message: %w", err)
		}
		defer os.Remove(tmpFile)

		// Run git commit
		err = utils.GitClient.Commit(commitMsg)
		if err != nil {
			return fmt.Errorf("failed to create commit: %w", err)
		}

		// Get the commit hash
		commitHash, err := utils.GitClient.GetCurrentHead()
		if err != nil {
			return fmt.Errorf("failed to get commit hash: %w", err)
		}

		// Get the author
		author, err := utils.GitClient.GetAuthor()
		if err != nil {
			return fmt.Errorf("failed to get author: %w", err)
		}

		// If we have a current story, associate this commit with it
		cfg, err := config.LoadConfig()
		if err == nil && cfg.JiraHost != "" && cfg.JiraProject != "" {
			storyID, err := utils.GitClient.GetConfig(fmt.Sprintf("%s.current.story", cfg.JiraProject))
			if err == nil && storyID != "" {
				// Load the story
				s, err := story.LoadStory(storyID)
				if err == nil {
					// Add the commit to the story
					s.AddCommit(commitHash, commitMsg, author, time.Now())

					// Get changed files
					files, err := utils.GitClient.GetChangedFiles()
					if err == nil {
						for _, file := range files {
							s.AddFile(file, "M") // Assuming modified for now
						}
					}

					// Save the updated story
					if err := s.Save(); err != nil {
						return fmt.Errorf("failed to update story: %w", err)
					}
				}
			}
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Created commit: %s\n", commitMsg)
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
