package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

	// First configure a project and user (required for story command)
	err = configureProject("test-project")
	require.NoError(t, err)
	err = configureUser("john.doe")
	require.NoError(t, err)

	tests := []struct {
		name        string
		title       string
		description string
		tags        []string
		number      int
		expectError bool
	}{
		{
			name:        "create story with number only",
			number:      123,
			expectError: false,
		},
		{
			name:        "create story with number and title",
			title:       "Test Story",
			number:      124,
			expectError: false,
		},
		{
			name:        "create story with number and description",
			description: "This is a test story",
			number:      125,
			expectError: false,
		},
		{
			name:        "create story with number and tags",
			tags:        []string{"test", "feature"},
			number:      126,
			expectError: false,
		},
		{
			name:        "create story without number",
			title:       "Test Story",
			expectError: true,
		},
		{
			name:        "create story with invalid number",
			title:       "Test Story",
			number:      0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up stories directory before each test
			storiesDir, err := story.GetStoriesDir()
			require.NoError(t, err)
			files, err := os.ReadDir(storiesDir)
			require.NoError(t, err)
			for _, file := range files {
				err = os.Remove(filepath.Join(storiesDir, file.Name()))
				require.NoError(t, err)
			}

			// Create a new command instance for each test
			cmd := &cobra.Command{
				Use:   "new",
				Short: "Create a new story",
				Long:  `Create a new story with title, description, and other metadata.`,
				RunE:  storyNewCmd.RunE,
			}

			// Add flags to the command
			cmd.Flags().StringP("title", "t", "", "Story title")
			cmd.Flags().StringP("description", "d", "", "Story description")
			cmd.Flags().StringSlice("tags", []string{}, "Story tags")
			cmd.Flags().IntP("number", "n", 0, "Story number")
			if err := cmd.MarkFlagRequired("number"); err != nil {
				t.Fatalf("failed to mark number flag as required: %v", err)
			}

			// Create a buffer to capture output
			var buf, errBuf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&errBuf) // Capture error output

			// Build command arguments
			args := []string{"--number", strconv.Itoa(tt.number)}
			if tt.title != "" {
				args = append(args, "--title", tt.title)
			}
			if tt.description != "" {
				args = append(args, "--description", tt.description)
			}
			if len(tt.tags) > 0 {
				args = append(args, "--tags", strings.Join(tt.tags, ","))
			}

			// Set command arguments
			cmd.SetArgs(args)

			err = cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify story was created
			stories, err := story.ListStories()
			require.NoError(t, err)
			require.Len(t, stories, 1, "Expected exactly one story to be created")

			// Get the created story
			createdStory := stories[0]
			assert.NotEmpty(t, createdStory.ID, "Story ID should not be empty")
			assert.Equal(t, tt.title, createdStory.Title, "Story title should match")
			assert.Equal(t, tt.description, createdStory.Description, "Story description should match")
			assert.Equal(t, tt.number, createdStory.Number, "Story number should match")
			assert.Equal(t, "john.doe", createdStory.Author, "Story author should match configured user")
			assert.Equal(t, "open", createdStory.Status, "Story status should be 'open'")
			if len(tt.tags) > 0 {
				assert.Equal(t, tt.tags, createdStory.Tags, "Story tags should match")
			}

			// Verify output format
			output := buf.String()
			expectedOutput := fmt.Sprintf("Created new story: %s\nNumber: %d\n", createdStory.ID, tt.number)
			if tt.title != "" {
				expectedOutput += fmt.Sprintf("Title: %s\n", tt.title)
			}
			if tt.description != "" {
				expectedOutput += fmt.Sprintf("Description: %s\n", tt.description)
			}
			if len(tt.tags) > 0 {
				expectedOutput += fmt.Sprintf("Tags: %v\n", tt.tags)
			}
			expectedOutput += "Author: john.doe\nStatus: open\n"
			assert.Equal(t, expectedOutput, output, "Command output should match expected format")
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
