package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/helmedeiros/tracer-bullet/internal/config"
	"github.com/helmedeiros/tracer-bullet/internal/story"
	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T) string {
	// Create a temporary directory for testing
	dir := t.TempDir()

	// Save current directory
	currentDir, err := os.Getwd()
	require.NoError(t, err)

	// Change to test directory
	err = os.Chdir(dir)
	require.NoError(t, err)

	// Initialize git repository using existing helper
	err = utils.RunGitInit()
	require.NoError(t, err)

	// Return to original directory
	err = os.Chdir(currentDir)
	require.NoError(t, err)

	return dir
}

func TestStoryCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		expectError    bool
		errorMsg       string
		expectedOutput string
	}{
		{
			name: "create story with number only",
			flags: map[string]string{
				"number": "123",
			},
			expectError: true,
			errorMsg:    "title is required when creating a story with a number",
		},
		{
			name: "create story with number and title",
			flags: map[string]string{
				"number": "123",
				"title":  "Test Story",
			},
			expectError: false,
		},
		{
			name: "create story with number and description",
			flags: map[string]string{
				"number":      "123",
				"title":       "Test Story",
				"description": "Test Description",
			},
			expectError: false,
		},
		{
			name: "create story with number and tags",
			flags: map[string]string{
				"number": "123",
				"title":  "Test Story",
				"tags":   "tag1,tag2",
			},
			expectError: false,
		},
		{
			name: "create story without number",
			flags: map[string]string{
				"title":       "Test Story",
				"description": "Test Description",
				"tags":        "tag1,tag2",
			},
			expectError: true,
			errorMsg:    "required flag(s) \"number\" not set",
		},
		{
			name: "create story with invalid number",
			flags: map[string]string{
				"number": "0",
				"title":  "Test Story",
			},
			expectError: true,
			errorMsg:    "number must be greater than 0",
		},
		{
			name: "create story without project configuration",
			flags: map[string]string{
				"number": "123",
				"title":  "Test Story",
			},
			expectError: true,
			errorMsg:    "project not configured. Please run 'tracer configure project' first",
		},
		{
			name: "create story without user configuration",
			flags: map[string]string{
				"number": "123",
				"title":  "Test Story",
			},
			expectError: true,
			errorMsg:    "user not configured. Please run 'tracer configure user' first",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test environment
			tmpDir, _, originalDir := setupTestEnvironment(t)
			defer cleanupTestEnvironment(t, tmpDir, originalDir)

			// Configure project and user for tests that need them
			if tt.name != "create story without project configuration" && tt.name != "create story without user configuration" {
				err := configureProject("test-project")
				require.NoError(t, err)
				err = configureUser("john.doe")
				require.NoError(t, err)
			} else if tt.name == "create story without user configuration" {
				err := configureProject("test-project")
				require.NoError(t, err)
			}

			// Create command
			rootCmd := &cobra.Command{Use: "tracer"}
			cmd := &cobra.Command{
				Use:   "new",
				Short: "Create a new story",
				Long:  `Create a new story with title, description, and other metadata.`,
				RunE:  storyNewCmd.RunE,
			}
			storyCmd := &cobra.Command{
				Use:   "story",
				Short: "Manage stories and their tracking",
			}
			storyCmd.AddCommand(cmd)
			rootCmd.AddCommand(storyCmd)

			// Initialize flags
			cmd.Flags().StringP("title", "t", "", "Story title")
			cmd.Flags().StringP("description", "d", "", "Story description")
			cmd.Flags().StringSliceP("tags", "g", []string{}, "Story tags")
			cmd.Flags().IntP("number", "n", 0, "Story number")
			if err := cmd.MarkFlagRequired("number"); err != nil {
				t.Fatalf("failed to mark number flag as required: %v", err)
			}

			// Set flags
			var err error
			if number, ok := tt.flags["number"]; ok && number != "" {
				err = cmd.Flags().Set("number", number)
				require.NoError(t, err)
			}
			if title, ok := tt.flags["title"]; ok {
				err = cmd.Flags().Set("title", title)
				require.NoError(t, err)
			}
			if description, ok := tt.flags["description"]; ok {
				err = cmd.Flags().Set("description", description)
				require.NoError(t, err)
			}
			if tags, ok := tt.flags["tags"]; ok {
				err = cmd.Flags().Set("tags", tags)
				require.NoError(t, err)
			}

			// Execute command
			rootCmd.SetArgs([]string{"story", "new"})
			err = rootCmd.Execute()
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Equal(t, tt.errorMsg, err.Error())
				}
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestStoryAfterHashCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid_hash",
			args:    []string{"--hash", "abc123"},
			wantErr: false,
		},
		{
			name:    "invalid_hash",
			args:    []string{"--hash", ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new command for each test
			cmd := &cobra.Command{
				Use:   "story",
				Short: "Manage stories",
			}
			afterHashCmd := &cobra.Command{
				Use:   "after-hash",
				Short: "Show stories after a specific commit hash",
				RunE: func(cmd *cobra.Command, args []string) error {
					hash, _ := cmd.Flags().GetString("hash")
					if hash == "" {
						return fmt.Errorf("hash is required")
					}
					return nil
				},
			}
			afterHashCmd.Flags().String("hash", "", "Commit hash")
			cmd.AddCommand(afterHashCmd)

			// Set the args
			cmd.SetArgs(append([]string{"after-hash"}, tt.args...))

			// Execute the command
			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("TestStoryAfterHashCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStoryByCommand(t *testing.T) {
	tmpDir, _, originalDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, tmpDir, originalDir)

	// First configure a project and user
	err := configureProject("test-project")
	require.NoError(t, err)
	err = configureUser("john.doe")
	require.NoError(t, err)

	// Create stories directory
	storyDir := filepath.Join(tmpDir, "stories")
	err = os.MkdirAll(storyDir, utils.DefaultDirPerm)
	require.NoError(t, err)

	// Update config with story directory
	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	cfg.StoryDir = storyDir
	err = config.SaveConfig(cfg)
	require.NoError(t, err)

	// Create test stories
	story1, err := story.NewStory("Story 1", "Description 1", "john.doe")
	require.NoError(t, err)
	story1.CreatedAt = time.Now().Add(-48 * time.Hour)
	story1.UpdatedAt = time.Now()
	story1.Commits = []story.Commit{
		{Hash: "abc123", Message: "Initial commit", Author: "john.doe", Timestamp: time.Now().Add(-24 * time.Hour)},
		{Hash: "def456", Message: "Second commit", Author: "john.doe", Timestamp: time.Now()},
	}
	story1.Files = []story.File{
		{Path: "file1.txt", Status: "added", Timestamp: time.Now().Add(-24 * time.Hour)},
		{Path: "file2.txt", Status: "modified", Timestamp: time.Now()},
	}
	err = story1.Save()
	require.NoError(t, err)

	story2, err := story.NewStory("Story 2", "Description 2", "jane.doe")
	require.NoError(t, err)
	err = story2.Save()
	require.NoError(t, err)

	tests := []struct {
		name        string
		author      string
		expectError bool
		expectCount int
	}{
		{
			name:        "valid author",
			author:      "john.doe",
			expectError: false,
			expectCount: 1,
		},
		{
			name:        "non-existent author",
			author:      "nonexistent",
			expectError: false,
			expectCount: 0,
		},
		{
			name:        "empty author",
			author:      "",
			expectError: true,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   "by",
				Short: "Show stories by author",
				RunE:  storyByCmd.RunE,
			}

			cmd.Flags().StringP("author", "a", "", "Story author")
			if err := cmd.MarkFlagRequired("author"); err != nil {
				t.Fatalf("failed to mark author flag as required: %v", err)
			}

			var buf, errBuf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&errBuf)

			cmd.SetArgs([]string{"--author", tt.author})

			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			output := buf.String()
			if tt.expectCount == 0 {
				assert.Contains(t, output, fmt.Sprintf("No stories found for author %s", tt.author))
			} else {
				assert.Contains(t, output, fmt.Sprintf("Stories by %s:", tt.author))
				assert.Contains(t, output, story1.Title)
				assert.Contains(t, output, story1.ID)
			}
		})
	}
}

func TestStoryDiaryCommand(t *testing.T) {
	// Set up test repository
	dir := setupTestRepo(t)

	// Save current directory
	currentDir, err := os.Getwd()
	require.NoError(t, err)

	// Change to test directory
	err = os.Chdir(dir)
	require.NoError(t, err)

	// Defer changing back to original directory
	defer func() {
		err := os.Chdir(currentDir)
		require.NoError(t, err)
	}()

	// First configure a project and user
	err = configureProject("test-project")
	require.NoError(t, err)
	err = configureUser("john.doe")
	require.NoError(t, err)

	// Create test story with commits and files
	now := time.Now()
	story1, err := story.NewStory("Story 1", "Description 1", "john.doe")
	require.NoError(t, err)
	story1.CreatedAt = now.Add(-48 * time.Hour)
	story1.UpdatedAt = now
	story1.Commits = []story.Commit{
		{Hash: "abc123", Message: "Initial commit", Author: "john.doe", Timestamp: now.Add(-24 * time.Hour)},
		{Hash: "def456", Message: "Second commit", Author: "john.doe", Timestamp: now},
	}
	story1.Files = []story.File{
		{Path: "file1.txt", Status: "added", Timestamp: now.Add(-24 * time.Hour)},
		{Path: "file2.txt", Status: "modified", Timestamp: now},
	}
	err = story1.Save()
	require.NoError(t, err)

	tests := []struct {
		name        string
		storyID     string
		since       string
		until       string
		expectError bool
		expectCount int
	}{
		{
			name:        "valid story with time range",
			storyID:     story1.Filename,
			since:       now.Add(-48 * time.Hour).Format(time.RFC3339),
			until:       now.Format(time.RFC3339),
			expectError: false,
			expectCount: 2,
		},
		{
			name:        "valid story without time range",
			storyID:     story1.Filename,
			expectError: false,
			expectCount: 2,
		},
		{
			name:        "invalid story ID",
			storyID:     "nonexistent",
			expectError: true,
			expectCount: 0,
		},
		{
			name:        "empty story ID",
			storyID:     "",
			expectError: true,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   "diary",
				Short: "Show story development diary",
				RunE:  storyDiaryCmd.RunE,
			}

			cmd.Flags().StringP("id", "i", "", "Story ID")
			cmd.Flags().String("since", "", "Start time (RFC3339 format)")
			cmd.Flags().String("until", "", "End time (RFC3339 format)")
			if err := cmd.MarkFlagRequired("id"); err != nil {
				t.Fatalf("failed to mark id flag as required: %v", err)
			}

			var buf, errBuf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&errBuf)

			args := []string{"--id", tt.storyID}
			if tt.since != "" {
				args = append(args, "--since", tt.since)
			}
			if tt.until != "" {
				args = append(args, "--until", tt.until)
			}
			cmd.SetArgs(args)

			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			output := buf.String()
			assert.Contains(t, output, "Story Diary:")
			assert.Contains(t, output, story1.Title)
			assert.Contains(t, output, story1.ID)
			if tt.expectCount > 0 {
				assert.Contains(t, output, "Commits:")
				assert.Contains(t, output, "File Changes:")
			}
		})
	}
}

func TestStoryDiffCommand(t *testing.T) {
	// Set up test repository
	dir := setupTestRepo(t)

	// Save current directory
	currentDir, err := os.Getwd()
	require.NoError(t, err)

	// Change to test directory
	err = os.Chdir(dir)
	require.NoError(t, err)

	// Defer changing back to original directory
	defer func() {
		err := os.Chdir(currentDir)
		require.NoError(t, err)
	}()

	// First configure a project and user
	err = configureProject("test-project")
	require.NoError(t, err)
	err = configureUser("john.doe")
	require.NoError(t, err)

	// Create test story with commits and files
	now := time.Now()
	story1, err := story.NewStory("Story 1", "Description 1", "john.doe")
	require.NoError(t, err)
	story1.CreatedAt = now.Add(-48 * time.Hour)
	story1.UpdatedAt = now
	story1.Commits = []story.Commit{
		{Hash: "abc123", Message: "Initial commit", Author: "john.doe", Timestamp: now.Add(-24 * time.Hour)},
		{Hash: "def456", Message: "Second commit", Author: "john.doe", Timestamp: now},
	}
	story1.Files = []story.File{
		{Path: "file1.txt", Status: "added", Timestamp: now.Add(-24 * time.Hour)},
		{Path: "file2.txt", Status: "modified", Timestamp: now},
	}
	err = story1.Save()
	require.NoError(t, err)

	tests := []struct {
		name        string
		storyID     string
		from        string
		to          string
		expectError bool
		expectCount int
	}{
		{
			name:        "valid story with time range",
			storyID:     story1.Filename,
			from:        now.Add(-48 * time.Hour).Format(time.RFC3339),
			to:          now.Format(time.RFC3339),
			expectError: false,
			expectCount: 25, // Header (3) + Commits section (1 + 2*8) + Files section (1 + 2*4)
		},
		{
			name:        "valid story without time range",
			storyID:     story1.Filename,
			expectError: false,
			expectCount: 25, // Same as above
		},
		{
			name:        "invalid story ID",
			storyID:     "nonexistent",
			expectError: true,
			expectCount: 0,
		},
		{
			name:        "empty story ID",
			storyID:     "",
			expectError: true,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   "diff",
				Short: "Show story changes",
				RunE:  storyDiffCmd.RunE,
			}

			cmd.Flags().StringP("id", "i", "", "Story ID")
			cmd.Flags().String("from", "", "Start point (RFC3339 format)")
			cmd.Flags().String("to", "", "End point (RFC3339 format)")
			if err := cmd.MarkFlagRequired("id"); err != nil {
				t.Fatalf("failed to mark id flag as required: %v", err)
			}

			var buf bytes.Buffer
			cmd.SetOut(&buf)

			args := []string{"--id", tt.storyID}
			if tt.from != "" {
				args = append(args, "--from", tt.from)
			}
			if tt.to != "" {
				args = append(args, "--to", tt.to)
			}
			cmd.SetArgs(args)

			err := cmd.Execute()

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				output := buf.String()
				lines := strings.Split(strings.TrimSpace(output), "\n")
				require.Equal(t, tt.expectCount, len(lines))
			}
		})
	}
}
