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
	Short: "Create and manage commits",
	Long: `Manage your commits through a natural workflow:

1. Create Commits
   tracer commit create --type feat --message "Add new feature"
   tracer commit create --type fix --message "Fix bug" --scope core

2. View History
   tracer commit show --id <commit-hash>
   tracer commit list --story <story-id>

3. Search and Filter
   tracer commit by --author <author>
   tracer commit since --date <date>

Each command follows a natural progression, helping you:
- Create well-structured commits
- Track changes effectively
- Maintain clear history
- Link commits to stories`,
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
	Long: `Create a new commit with a conventional commit message.

Examples:
  tracer commit create --type feat --message "Add user authentication"
  tracer commit create --type fix --scope api --message "Fix timeout issue"
  tracer commit create --type feat --message "Breaking change" --breaking
  tracer commit create --auto  # Automatically generate commit message from changes

Commit Types:
  feat     - A new feature
  fix      - A bug fix
  docs     - Documentation only changes
  style    - Changes that don't affect the code's meaning
  refactor - Code changes that neither fix a bug nor add a feature
  test     - Adding missing tests or correcting existing tests
  chore    - Changes to the build process or auxiliary tools

Flags:
  --type     Required. The type of change (feat, fix, etc.)
  --message  Required. The commit message
  --scope    Optional. The scope of the change (e.g., api, core)
  --body     Optional. Detailed description of the change
  --breaking Optional. Mark as a breaking change
  --jira    Optional. Include Jira story URL in commit body
  --auto    Optional. Automatically generate commit message from changes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flag values
		commitType, _ := cmd.Flags().GetString("type")
		scope, _ := cmd.Flags().GetString("scope")
		message, _ := cmd.Flags().GetString("message")
		body, _ := cmd.Flags().GetString("body")
		breaking, _ := cmd.Flags().GetBool("breaking")
		includeJira, _ := cmd.Flags().GetBool("jira")
		auto, _ := cmd.Flags().GetBool("auto")

		// If auto flag is set, generate commit message from changes
		if auto {
			// Get unstaged and untracked files
			unstagedFiles, err := utils.GitClient.GetUnstagedFiles()
			if err != nil {
				return fmt.Errorf("failed to get unstaged files: %w", err)
			}

			untrackedFiles, err := utils.GitClient.GetUntrackedFiles()
			if err != nil {
				return fmt.Errorf("failed to get untracked files: %w", err)
			}

			if len(unstagedFiles) == 0 && len(untrackedFiles) == 0 {
				return fmt.Errorf("no changes to commit")
			}

			// Get diffs for all files
			var diffs []string
			for _, file := range unstagedFiles {
				diff, err := utils.GitClient.GetDiff(file)
				if err != nil {
					return fmt.Errorf("failed to get diff for %s: %w", file, err)
				}
				diffs = append(diffs, fmt.Sprintf("File: %s\n%s", file, diff))
			}

			for _, file := range untrackedFiles {
				content, err := os.ReadFile(file)
				if err != nil {
					return fmt.Errorf("failed to read untracked file %s: %w", file, err)
				}
				diffs = append(diffs, fmt.Sprintf("New file: %s\n%s", file, string(content)))
			}

			// Use llama to generate commit message
			commitMsg, err := utils.GenerateCommitMessage(diffs)
			if err != nil {
				return fmt.Errorf("failed to generate commit message: %w", err)
			}

			// Create temporary file for commit message
			tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("tracer-commit-msg-%d", time.Now().UnixNano()))
			if err := os.WriteFile(tmpFile, []byte(commitMsg), 0600); err != nil {
				return fmt.Errorf("failed to write commit message: %w", err)
			}
			defer os.Remove(tmpFile)

			// Stage all changes
			if err := utils.GitClient.StageAll(); err != nil {
				return fmt.Errorf("failed to stage changes: %w", err)
			}

			// Create commit using the temporary file
			if err := utils.GitClient.CommitWithFile(tmpFile); err != nil {
				return fmt.Errorf("failed to create commit: %w", err)
			}

			// Get the commit hash
			commitHash, err := utils.GitClient.GetCurrentHead()
			if err != nil {
				return fmt.Errorf("failed to get commit hash: %w", err)
			}

			// Display success message
			fmt.Fprintf(cmd.OutOrStdout(), "\nCommit created successfully!\n\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Details:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  Hash: %s\n", commitHash)
			fmt.Fprintf(cmd.OutOrStdout(), "  Message:\n%s\n", commitMsg)

			return nil
		}

		// Validate commit type with better error message
		if err := validateCommitType(commitType); err != nil {
			return fmt.Errorf(`%w

Valid commit types:
  feat     - A new feature
  fix      - A bug fix
  docs     - Documentation only changes
  style    - Changes that don't affect code meaning
  refactor - Code changes (no fixes/features)
  test     - Changes to tests
  chore    - Changes to build process/tools

Example:
  tracer commit create --type feat --message "Add user auth"`, err)
		}

		// Validate message with better guidance
		if message == "" {
			return fmt.Errorf(`commit message cannot be empty

Example messages:
  "Add user authentication flow"
  "Fix timeout in API requests"
  "Update documentation for setup"

Usage:
  tracer commit create --type <type> --message "Your message"`)
		}

		// Build commit message
		commitMsg := buildCommitMessage(commitType, scope, message, body, breaking)

		// Add Jira story URL if requested
		if includeJira {
			var err error
			commitMsg, err = addJiraUrl(commitMsg)
			if err != nil {
				return fmt.Errorf("failed to add Jira URL: %w", err)
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
			return fmt.Errorf(`failed to create commit: %w

Common issues:
1. No staged changes (use 'git add' first)
2. No changes to commit
3. Merge conflicts need resolution

Try these steps:
1. Stage your changes:   git add <files>
2. Check status:         git status
3. Try commit again:     tracer commit create ...`, err)
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

		// Display success message with next steps
		fmt.Fprintf(cmd.OutOrStdout(), "\nCommit created successfully!\n\n")
		fmt.Fprintf(cmd.OutOrStdout(), "Details:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  Type: %s\n", commitType)
		if scope != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  Scope: %s\n", scope)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  Message: %s\n", message)
		fmt.Fprintf(cmd.OutOrStdout(), "  Hash: %s\n", commitHash)
		fmt.Fprintf(cmd.OutOrStdout(), "  Author: %s\n", author)

		fmt.Fprintf(cmd.OutOrStdout(), "\nNext steps:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "1. Push changes:        git push\n")
		fmt.Fprintf(cmd.OutOrStdout(), "2. View commit:         git show %s\n", commitHash)
		fmt.Fprintf(cmd.OutOrStdout(), "3. Create more commits: tracer commit create ...\n")
		fmt.Fprintf(cmd.OutOrStdout(), "4. View story:          tracer story show\n")

		return nil
	},
}

var commitPreviewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Preview a commit message",
	Long: `Preview a commit message that would be generated from your changes.

Examples:
  tracer commit preview --auto  # Preview auto-generated commit message from changes

Flags:
  --auto    Optional. Automatically generate commit message from changes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flag values
		auto, _ := cmd.Flags().GetBool("auto")

		if !auto {
			return fmt.Errorf("preview command currently only supports --auto flag")
		}

		// Get unstaged and untracked files
		unstagedFiles, err := utils.GitClient.GetUnstagedFiles()
		if err != nil {
			return fmt.Errorf("failed to get unstaged files: %w", err)
		}

		untrackedFiles, err := utils.GitClient.GetUntrackedFiles()
		if err != nil {
			return fmt.Errorf("failed to get untracked files: %w", err)
		}

		if len(unstagedFiles) == 0 && len(untrackedFiles) == 0 {
			return fmt.Errorf("no changes to preview")
		}

		// Get diffs for all files
		var diffs []string
		for _, file := range unstagedFiles {
			diff, err := utils.GitClient.GetDiff(file)
			if err != nil {
				return fmt.Errorf("failed to get diff for %s: %w", file, err)
			}
			diffs = append(diffs, fmt.Sprintf("File: %s\n%s", file, diff))
		}

		for _, file := range untrackedFiles {
			content, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read untracked file %s: %w", file, err)
			}
			diffs = append(diffs, fmt.Sprintf("New file: %s\n%s", file, string(content)))
		}

		// Use llama to generate commit message
		commitMsg, err := utils.GenerateCommitMessage(diffs)
		if err != nil {
			return fmt.Errorf("failed to generate commit message: %w", err)
		}

		// Display the preview
		fmt.Fprintf(cmd.OutOrStdout(), "\nPreview of commit message:\n\n%s\n", commitMsg)
		fmt.Fprintf(cmd.OutOrStdout(), "\nTo create this commit, run:\n  tracer commit create --auto\n")

		return nil
	},
}

func init() {
	CommitCmd.AddCommand(commitCreateCmd)
	CommitCmd.AddCommand(commitPreviewCmd)

	// Add flags with better descriptions
	commitCreateCmd.Flags().String("type", "", "Type of change (feat, fix, docs, style, refactor, test, chore)")
	commitCreateCmd.Flags().String("message", "", "Short, descriptive commit message")
	commitCreateCmd.Flags().String("scope", "", "Scope of the change (e.g., api, core)")
	commitCreateCmd.Flags().String("body", "", "Detailed description of the change")
	commitCreateCmd.Flags().Bool("breaking", false, "Mark as a breaking change")
	commitCreateCmd.Flags().Bool("jira", false, "Include Jira story URL in commit body")
	commitCreateCmd.Flags().Bool("auto", false, "Automatically generate commit message from changes")

	// Add flags for preview command
	commitPreviewCmd.Flags().Bool("auto", false, "Automatically generate commit message from changes")

	// Mark required flags only if auto is not set
	commitCreateCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		auto, _ := cmd.Flags().GetBool("auto")
		if !auto {
			requiredFlags := []string{"type", "message"}
			for _, flag := range requiredFlags {
				if err := cmd.MarkFlagRequired(flag); err != nil {
					return fmt.Errorf("failed to mark %s flag as required: %v", flag, err)
				}
			}
		}
		return nil
	}
}
