package story

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/helmedeiros/tracer-bullet/internal/config"
	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewStory(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
		author      string
		wantErr     bool
	}{
		{
			name:        "valid story",
			title:       "Test Story",
			description: "This is a test story",
			author:      "test@example.com",
			wantErr:     false,
		},
		{
			name:        "empty title",
			title:       "",
			description: "This is a test story",
			author:      "test@example.com",
			wantErr:     true,
		},
		{
			name:        "empty author",
			title:       "Test Story",
			description: "This is a test story",
			author:      "",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			story, err := NewStory(tt.title, tt.description, tt.author)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, story)
			assert.Equal(t, tt.title, story.Title)
			assert.Equal(t, tt.description, story.Description)
			assert.Equal(t, tt.author, story.Author)
			assert.Equal(t, "open", story.Status)
			assert.NotEmpty(t, story.ID)
			assert.NotZero(t, story.CreatedAt)
			assert.NotZero(t, story.UpdatedAt)
			assert.Empty(t, story.Tags)
		})
	}
}

func TestSaveAndLoadStory(t *testing.T) {
	// Create a temporary directory for testing
	dir := t.TempDir()
	utils.TestConfigDir = dir
	defer func() { utils.TestConfigDir = "" }()

	// Create default config
	cfg := config.DefaultConfig()
	cfg.StoryDir = filepath.Join(dir, "stories")
	err := config.SaveConfig(cfg)
	assert.NoError(t, err)

	// Create a test story
	story, err := NewStory("Test Story", "This is a test story", "test@example.com")
	assert.NoError(t, err)

	// Test saving the story
	err = story.Save()
	assert.NoError(t, err)

	// Verify the story file was created
	storyFile := filepath.Join(cfg.StoryDir, story.ID+".json")
	assert.FileExists(t, storyFile)

	// Test loading the story
	loadedStory, err := LoadStory(story.ID)
	assert.NoError(t, err)
	assert.NotNil(t, loadedStory)
	assert.Equal(t, story.ID, loadedStory.ID)
	assert.Equal(t, story.Title, loadedStory.Title)
	assert.Equal(t, story.Description, loadedStory.Description)
	assert.Equal(t, story.Author, loadedStory.Author)
	assert.Equal(t, story.Status, loadedStory.Status)
}

func TestListStories(t *testing.T) {
	// Create a temporary directory for testing
	dir := t.TempDir()
	utils.TestConfigDir = dir
	defer func() { utils.TestConfigDir = "" }()

	// Create default config
	cfg := config.DefaultConfig()
	cfg.StoryDir = filepath.Join(dir, "stories")
	err := config.SaveConfig(cfg)
	assert.NoError(t, err)

	// Create some test stories
	stories := []*Story{
		{ID: "story1", Title: "Story 1", CreatedAt: time.Now()},
		{ID: "story2", Title: "Story 2", CreatedAt: time.Now().Add(-time.Hour)},
		{ID: "story3", Title: "Story 3", CreatedAt: time.Now().Add(-2 * time.Hour)},
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
