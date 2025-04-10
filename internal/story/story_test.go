package story

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	// Save the original git client
	originalGitClient := utils.GitClient

	// Create and set up mock git client
	mockGit := utils.NewMockGit()
	utils.GitClient = mockGit.(*utils.MockGit)

	// Configure mock behavior
	mockGit.(*utils.MockGit).GetGitRootFunc = func() (string, error) {
		return dir, nil
	}
	mockGit.(*utils.MockGit).InitFunc = func() error {
		return nil
	}

	// Initialize mock git repository
	err = utils.RunGitInit()
	assert.NoError(t, err)

	// Return to original directory
	err = os.Chdir(currentDir)
	assert.NoError(t, err)

	// Restore the original git client when the test is done
	t.Cleanup(func() {
		utils.GitClient = originalGitClient
	})

	return dir
}

func TestNewStory(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
		author      string
		number      int
		project     string
		expectError bool
		errorMsg    string
		branchName  string
	}{
		{
			name:        "valid story with number and title and project",
			title:       "Test Story",
			description: "Test Description",
			author:      "john.doe",
			number:      123,
			project:     "my-project",
			expectError: false,
			branchName:  "features/my-project-123-test-story",
		},
		{
			name:        "valid story with all fields and project",
			title:       "Test Story",
			description: "Test Description",
			author:      "john.doe",
			number:      124,
			project:     "my-project",
			expectError: false,
			branchName:  "features/my-project-124-test-story",
		},
		{
			name:        "missing number",
			title:       "Test Story",
			description: "Test Description",
			author:      "john.doe",
			number:      0,
			project:     "my-project",
			expectError: true,
			errorMsg:    "number must be greater than 0",
		},
		{
			name:        "missing title with number",
			title:       "",
			description: "Test Description",
			author:      "john.doe",
			number:      125,
			project:     "my-project",
			expectError: true,
			errorMsg:    "title is required when creating a story with a number",
		},
		{
			name:        "story with special characters and project",
			title:       "Test Story: Fix Bug #123",
			description: "Test Description",
			author:      "john.doe",
			number:      126,
			project:     "my-project",
			expectError: false,
			branchName:  "features/my-project-126-test-story-fix-bug-123",
		},
		{
			name:        "valid story with number and title without project",
			title:       "Test Story",
			description: "Test Description",
			author:      "john.doe",
			number:      127,
			project:     "",
			expectError: false,
			branchName:  "features/127-test-story",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mock git client
			originalGitClient := utils.GitClient
			defer func() {
				utils.GitClient = originalGitClient
			}()

			mockGit := utils.NewMockGit()
			utils.GitClient = mockGit.(*utils.MockGit)

			// Configure mock behavior
			mockGit.(*utils.MockGit).BranchExistsFunc = func(branchName string) (bool, error) {
				return false, nil
			}
			mockGit.(*utils.MockGit).CreateBranchFunc = func(branchName string) error {
				if tt.branchName != "" {
					assert.Equal(t, tt.branchName, branchName)
				}
				return nil
			}
			mockGit.(*utils.MockGit).GetConfigFunc = func(key string) (string, error) {
				if key == "current.project" {
					return tt.project, nil
				}
				return "", nil
			}

			story, err := NewStoryWithNumber(tt.title, tt.description, tt.author, tt.number)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Equal(t, tt.errorMsg, err.Error())
				}
				return
			}
			require.NoError(t, err)
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

func TestNewStory_BranchExists(t *testing.T) {
	// Set up mock git client
	originalGitClient := utils.GitClient
	defer func() {
		utils.GitClient = originalGitClient
	}()

	mockGit := utils.NewMockGit()
	utils.GitClient = mockGit.(*utils.MockGit)

	// Configure mock behavior
	mockGit.(*utils.MockGit).BranchExistsFunc = func(branchName string) (bool, error) {
		return true, nil
	}
	mockGit.(*utils.MockGit).SwitchBranchFunc = func(branchName string) error {
		assert.Equal(t, "features/my-project-123-test-story", branchName)
		return nil
	}
	mockGit.(*utils.MockGit).GetConfigFunc = func(key string) (string, error) {
		if key == "current.project" {
			return "my-project", nil
		}
		return "", nil
	}

	// Create a story
	story, err := NewStoryWithNumber("Test Story", "Test Description", "john.doe", 123)
	require.NoError(t, err)
	assert.NotNil(t, story)
}

func TestNewStory_NoGitRepo(t *testing.T) {
	// Set up mock git client
	originalGitClient := utils.GitClient
	defer func() {
		utils.GitClient = originalGitClient
	}()

	mockGit := utils.NewMockGit()
	utils.GitClient = mockGit.(*utils.MockGit)

	// Configure mock behavior to simulate not being in a git repo
	mockGit.(*utils.MockGit).BranchExistsFunc = func(branchName string) (bool, error) {
		return false, fmt.Errorf("not a git repository")
	}

	// Create a story - should still work even though git operations fail
	story, err := NewStory("Test Story", "Test Description", "john.doe")
	require.NoError(t, err)
	assert.NotNil(t, story)
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

	// Get stories directory
	storiesDir, err := GetStoriesDir()
	assert.NoError(t, err)

	// Clean up any existing stories
	err = os.RemoveAll(storiesDir)
	assert.NoError(t, err)
	err = os.MkdirAll(storiesDir, 0755)
	assert.NoError(t, err)

	// Create test stories with proper initialization
	now := time.Now()
	stories := []*Story{
		{
			ID:          "story1",
			Title:       "Story 1",
			Description: "Description 1",
			Status:      "open",
			CreatedAt:   now,
			UpdatedAt:   now,
			Author:      "test-author",
			Tags:        []string{},
			Filename:    "story1.yaml",
		},
		{
			ID:          "story2",
			Title:       "Story 2",
			Description: "Description 2",
			Status:      "open",
			CreatedAt:   now.Add(-time.Hour),
			UpdatedAt:   now.Add(-time.Hour),
			Author:      "test-author",
			Tags:        []string{},
			Filename:    "story2.yaml",
		},
		{
			ID:          "story3",
			Title:       "Story 3",
			Description: "Description 3",
			Status:      "open",
			CreatedAt:   now.Add(-2 * time.Hour),
			UpdatedAt:   now.Add(-2 * time.Hour),
			Author:      "test-author",
			Tags:        []string{},
			Filename:    "story3.yaml",
		},
	}

	// Save the stories
	for _, story := range stories {
		err := story.Save()
		assert.NoError(t, err)
	}

	// Test listing stories
	loadedStories, err := ListStories()
	assert.NoError(t, err)
	assert.Equal(t, len(stories), len(loadedStories), "Number of stories should match")

	// Verify the stories are sorted by creation date (newest first)
	for i := 0; i < len(loadedStories)-1; i++ {
		assert.True(t, loadedStories[i].CreatedAt.After(loadedStories[i+1].CreatedAt),
			"Stories should be sorted by creation date (newest first)")
	}

	// Verify story contents
	for _, story := range stories {
		// Find the corresponding loaded story
		var loadedStory *Story
		for _, ls := range loadedStories {
			if ls.ID == story.ID {
				loadedStory = ls
				break
			}
		}
		assert.NotNil(t, loadedStory, "Story %s should be found in loaded stories", story.ID)
		assert.Equal(t, story.Title, loadedStory.Title)
		assert.Equal(t, story.Description, loadedStory.Description)
		assert.Equal(t, story.Status, loadedStory.Status)
		assert.Equal(t, story.Author, loadedStory.Author)
	}
}
