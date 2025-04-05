package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCommitCommand(t *testing.T) {
	// Save the original config dir
	origConfigDir := utils.TestConfigDir
	defer func() {
		utils.TestConfigDir = origConfigDir
	}()

	tests := []struct {
		name        string
		commitType  string
		message     string
		scope       string
		body        string
		breaking    bool
		jira        bool
		expectError bool
		expectedMsg string
	}{
		{
			name:        "basic commit",
			commitType:  "feat",
			message:     "add new feature",
			expectError: false,
			expectedMsg: "feat: add new feature",
		},
		{
			name:        "commit with scope",
			commitType:  "fix",
			message:     "fix bug",
			scope:       "core",
			expectError: false,
			expectedMsg: "fix(core): fix bug",
		},
		{
			name:        "commit with body",
			commitType:  "feat",
			message:     "add feature",
			body:        "This is a detailed description",
			expectError: false,
			expectedMsg: "feat: add feature\n\nThis is a detailed description",
		},
		{
			name:        "breaking change",
			commitType:  "feat",
			message:     "breaking change",
			breaking:    true,
			expectError: false,
			expectedMsg: "feat!: breaking change",
		},
		{
			name:        "invalid commit type",
			commitType:  "invalid",
			message:     "test message",
			expectError: true,
			expectedMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir := t.TempDir()
			repoDir := filepath.Join(tmpDir, "repo")
			err := os.MkdirAll(repoDir, 0755)
			assert.NoError(t, err)

			// Set up test config directory
			utils.TestConfigDir = filepath.Join(repoDir, ".tracer")
			err = os.MkdirAll(utils.TestConfigDir, 0755)
			assert.NoError(t, err)

			// Save current directory
			currentDir, err := os.Getwd()
			assert.NoError(t, err)

			// Change to test directory
			err = os.Chdir(repoDir)
			assert.NoError(t, err)
			defer func() {
				err = os.Chdir(currentDir)
				assert.NoError(t, err)
			}()

			// Create a mock git client
			mockGit := utils.NewMockGit()
			utils.GitClient = mockGit

			// Configure mock behavior
			mockGit.(*utils.MockGit).SetConfigFunc = func(key, value string) error {
				return nil
			}
			mockGit.(*utils.MockGit).GetConfigFunc = func(key string) (string, error) {
				switch key {
				case "current.project":
					return "test-project", nil
				case "user.name":
					return "Test User", nil
				default:
					return "", nil
				}
			}
			mockGit.(*utils.MockGit).GetAuthorFunc = func() (string, error) {
				return "Test User", nil
			}
			mockGit.(*utils.MockGit).CommitFunc = func(message string) error {
				return nil
			}
			mockGit.(*utils.MockGit).GetChangedFilesFunc = func() ([]string, error) {
				return []string{"test.txt"}, nil
			}
			mockGit.(*utils.MockGit).InitFunc = func() error {
				return nil
			}
			mockGit.(*utils.MockGit).GetCurrentHeadFunc = func() (string, error) {
				return "abc123", nil
			}
			mockGit.(*utils.MockGit).GetGitRootFunc = func() (string, error) {
				return repoDir, nil
			}

			// Create a test file
			err = os.WriteFile(filepath.Join(repoDir, "test.txt"), []byte("test content"), 0644)
			assert.NoError(t, err)

			// Initialize the root command and add the commit command
			rootCmd := &cobra.Command{Use: "tracer"}
			rootCmd.AddCommand(CommitCmd)

			// Set up the command arguments
			args := []string{
				"commit", "create",
				"--type", tt.commitType,
				"--message", tt.message,
			}
			if tt.scope != "" {
				args = append(args, "--scope", tt.scope)
			}
			if tt.body != "" {
				args = append(args, "--body", tt.body)
			}
			if tt.breaking {
				args = append(args, "--breaking")
			}
			if tt.jira {
				args = append(args, "--jira")
			}

			// Set the command's args
			rootCmd.SetArgs(args)

			// Execute the command
			err = rootCmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid commit type")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
