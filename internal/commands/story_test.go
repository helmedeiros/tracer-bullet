package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/helmedeiros/tracer-bullet/internal/config"
	"github.com/helmedeiros/tracer-bullet/internal/story"
	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoryCommand(t *testing.T) {
	tmpDir, originalDir := setupTestEnvironment(t)
	defer func() {
		err := os.Chdir(originalDir)
		require.NoError(t, err)
		os.RemoveAll(tmpDir)
	}()

	// First configure a project and user (required for story command)
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

	tests := []struct {
		name        string
		title       string
		description string
		tags        []string
		expectError bool
	}{
		{
			name:        "create story with title",
			title:       "Test Story",
			expectError: false,
		},
		{
			name:        "create story with title and description",
			title:       "Another Story",
			description: "This is a test story",
			expectError: false,
		},
		{
			name:        "create story with title and tags",
			title:       "Tagged Story",
			tags:        []string{"test", "feature"},
			expectError: false,
		},
		{
			name:        "create story without title",
			description: "Missing title",
			expectError: true,
		},
		{
			name:        "create story with empty title",
			title:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up stories directory before each test
			var files []os.DirEntry
			files, err = os.ReadDir(storyDir)
			require.NoError(t, err)
			for _, file := range files {
				err = os.Remove(filepath.Join(storyDir, file.Name()))
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
			cmd.Flags().StringP("title", "t", "", "Story title (required)")
			cmd.Flags().StringP("description", "d", "", "Story description")
			cmd.Flags().StringSlice("tags", []string{}, "Story tags")
			cmd.MarkFlagRequired("title")

			// Create a buffer to capture output
			var buf, errBuf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&errBuf) // Capture error output

			// Build command arguments
			args := []string{"--title", tt.title}
			if tt.description != "" {
				args = append(args, "--description", tt.description)
			}
			if len(tt.tags) > 0 {
				args = append(args, "--tags", strings.Join(tt.tags, ","))
			}

			// Set command arguments
			cmd.SetArgs(args)

			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				// Verify the error message is correct but don't print it
				assert.Contains(t, errBuf.String(), "title cannot be empty", "Error message should indicate title is required")
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
			assert.Equal(t, "john.doe", createdStory.Author, "Story author should match configured user")
			assert.Equal(t, "open", createdStory.Status, "Story status should be 'open'")
			if len(tt.tags) > 0 {
				assert.Equal(t, tt.tags, createdStory.Tags, "Story tags should match")
			}

			// Verify output format
			output := buf.String()
			expectedOutput := fmt.Sprintf("Created new story: %s\nTitle: %s\n", createdStory.ID, tt.title)
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
