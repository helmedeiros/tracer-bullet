package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCommitCommand(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "tracer-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a temporary git repository
	gitDir := filepath.Join(tmpDir, "test-repo")
	err = os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Initialize git repository
	err = os.Chdir(gitDir)
	if err != nil {
		t.Fatal(err)
	}

	// Run git init
	cmd := exec.Command("git", "init")
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	// Configure git user
	cmd = exec.Command("git", "config", "user.name", "Test User")
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	err = cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "create feat commit",
			args:    []string{"create", "--type", "feat", "--message", "add new feature"},
			wantErr: false,
		},
		{
			name:    "create feat commit with scope",
			args:    []string{"create", "--type", "feat", "--scope", "auth", "--message", "add login feature"},
			wantErr: false,
		},
		{
			name:    "create fix commit with body",
			args:    []string{"create", "--type", "fix", "--message", "fix bug", "--body", "This fixes a critical issue in the login flow"},
			wantErr: false,
		},
		{
			name:    "create breaking change",
			args:    []string{"create", "--type", "feat", "--message", "change API", "--breaking"},
			wantErr: false,
		},
		{
			name:        "invalid commit type",
			args:        []string{"create", "--type", "invalid", "--message", "test"},
			wantErr:     true,
			errContains: "invalid commit type",
		},
		{
			name:        "empty message",
			args:        []string{"create", "--type", "feat", "--message", ""},
			wantErr:     true,
			errContains: "commit message cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test file with unique content for each test
			testFile := filepath.Join(gitDir, "test.txt")
			err = os.WriteFile(testFile, []byte("test content for "+tt.name), 0644)
			if err != nil {
				t.Fatal(err)
			}

			// Add the test file
			cmd = exec.Command("git", "add", "test.txt")
			err = cmd.Run()
			if err != nil {
				t.Fatal(err)
			}

			// Reset the command and its flags
			CommitCmd = &cobra.Command{
				Use:   "commit",
				Short: "Create a conventional commit",
				Long: `Create a git commit following conventional commit format.
Supports common types like feat, fix, docs, style, refactor, test, and chore.
Automatically includes scope based on current story if available.`,
			}

			// Re-initialize the command
			commitCreateCmd = &cobra.Command{
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
					if err := os.WriteFile(tmpFile, []byte(commitMsg.String()), 0644); err != nil {
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

			// Add create command flags
			commitCreateCmd.Flags().String("type", "", "Commit type (feat, fix, docs, style, refactor, test, chore)")
			commitCreateCmd.Flags().String("scope", "", "Commit scope (optional)")
			commitCreateCmd.Flags().String("message", "", "Commit message")
			commitCreateCmd.Flags().String("body", "", "Commit body (optional)")
			commitCreateCmd.Flags().Bool("breaking", false, "Mark as breaking change")

			// Mark required flags
			if err := commitCreateCmd.MarkFlagRequired("type"); err != nil {
				t.Fatalf("failed to mark type flag as required: %v", err)
			}
			if err := commitCreateCmd.MarkFlagRequired("message"); err != nil {
				t.Fatalf("failed to mark message flag as required: %v", err)
			}

			// Add commands to root
			CommitCmd.AddCommand(commitCreateCmd)

			// Set the command arguments
			CommitCmd.SetArgs(tt.args)

			// Execute the command
			err := CommitCmd.Execute()

			// Check error cases
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			// Check success cases
			assert.NoError(t, err)

			// Verify the commit was created
			cmd := exec.Command("git", "log", "-1", "--pretty=format:%s")
			output, err := cmd.Output()
			if err != nil {
				t.Fatal(err)
			}

			// Verify commit message format based on test case
			switch tt.name {
			case "create feat commit":
				assert.Equal(t, "feat: add new feature", string(output))
			case "create feat commit with scope":
				assert.Equal(t, "feat(auth): add login feature", string(output))
			case "create fix commit with body":
				cmd = exec.Command("git", "log", "-1", "--pretty=format:%B")
				output, err = cmd.Output()
				if err != nil {
					t.Fatal(err)
				}
				assert.Contains(t, string(output), "This fixes a critical issue in the login flow")
			case "create breaking change":
				assert.Equal(t, "feat!: change API", string(output))
			}
		})
	}
}
