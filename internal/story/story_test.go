package story

import (
	"os"
	"testing"
	"time"

	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"github.com/stretchr/testify/assert"
)

func setupTestRepo(t *testing.T) string {
	// Create a temporary directory for testing
	dir := t.TempDir()

	// Save current directory
	currentDir, err := os.Getwd()
	assert.NoError(t, err)

	// Change to test directory
	err = os.Chdir(dir)
	assert.NoError(t, err)

	// Initialize git repository using existing helper
	err = utils.RunGitInit()
	assert.NoError(t, err)

	// Return to original directory
	err = os.Chdir(currentDir)
	assert.NoError(t, err)

	return dir
}

func TestNewStory(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
		author      string
		number      int
		wantErr     bool
	}{
		{
			name:    "valid story with number only",
			number:  123,
			wantErr: false,
		},
		{
			name:        "valid story with all fields",
			title:       "Test Story",
			description: "This is a test story",
			author:      "test@example.com",
			number:      124,
			wantErr:     false,
		},
		{
			name:        "missing number",
			title:       "Test Story",
			description: "This is a test story",
			author:      "test@example.com",
			number:      0,
			wantErr:     true,
		},
		{
			name:        "negative number",
			title:       "Test Story",
			description: "This is a test story",
			author:      "test@example.com",
			number:      -1,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			story, err := NewStoryWithNumber(tt.title, tt.description, tt.author, tt.number)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, story)
			assert.Equal(t, tt.title, story.Title)
			assert.Equal(t, tt.description, story.Description)
			assert.Equal(t, tt.author, story.Author)
			assert.Equal(t, tt.number, story.Number)
			assert.Equal(t, "open", story.Status)
			assert.NotEmpty(t, story.ID)
			assert.NotZero(t, story.CreatedAt)
			assert.NotZero(t, story.UpdatedAt)
			assert.Empty(t, story.Tags)
		})
	}
}

func TestSaveAndLoadStory(t *testing.T) {
	// Set up test repository
	dir := setupTestRepo(t)

	// Save current directory
	currentDir, err := os.Getwd()
	assert.NoError(t, err)

	// Change to test directory
	err = os.Chdir(dir)
	assert.NoError(t, err)

	// Defer changing back to original directory
	defer func() {
		err := os.Chdir(currentDir)
		assert.NoError(t, err)
	}()

	// Create a test story
	story, err := NewStory("Test Story", "This is a test story", "test@example.com")
	assert.NoError(t, err)

	// Test saving the story
	err = story.Save()
	assert.NoError(t, err)

	// Test loading the story
	loadedStory, err := LoadStory(story.Filename)
	assert.NoError(t, err)
	assert.NotNil(t, loadedStory)
	assert.Equal(t, story.ID, loadedStory.ID)
	assert.Equal(t, story.Title, loadedStory.Title)
	assert.Equal(t, story.Description, loadedStory.Description)
	assert.Equal(t, story.Author, loadedStory.Author)
	assert.Equal(t, story.Status, loadedStory.Status)
}

func TestListStories(t *testing.T) {
	// Set up test repository
	dir := setupTestRepo(t)

	// Save current directory
	currentDir, err := os.Getwd()
	assert.NoError(t, err)

	// Change to test directory
	err = os.Chdir(dir)
	assert.NoError(t, err)

	// Defer changing back to original directory
	defer func() {
		err := os.Chdir(currentDir)
		assert.NoError(t, err)
	}()

	// Create some test stories
	stories := []*Story{
		{ID: "story1", Title: "Story 1", CreatedAt: time.Now(), Filename: "story1.yaml"},
		{ID: "story2", Title: "Story 2", CreatedAt: time.Now().Add(-time.Hour), Filename: "story2.yaml"},
		{ID: "story3", Title: "Story 3", CreatedAt: time.Now().Add(-2 * time.Hour), Filename: "story3.yaml"},
	}

	// Save the stories
	for _, story := range stories {
		err := story.Save()
		assert.NoError(t, err)
	}

	// Test listing stories
	loadedStories, err := ListStories()
	assert.NoError(t, err)
	assert.Len(t, loadedStories, len(stories))

	// Verify the stories are sorted by creation date (newest first)
	for i := 0; i < len(stories)-1; i++ {
		assert.True(t, loadedStories[i].CreatedAt.After(loadedStories[i+1].CreatedAt))
	}
}
