package story

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"gopkg.in/yaml.v3"
)

// Commit represents a Git commit associated with a story
type Commit struct {
	Hash      string    `json:"hash"`
	Message   string    `json:"message"`
	Author    string    `json:"author"`
	Timestamp time.Time `json:"timestamp"`
}

// File represents a file modified as part of a story
type File struct {
	Path      string    `json:"path"`
	Status    string    `json:"status"` // added, modified, deleted
	Timestamp time.Time `json:"timestamp"`
}

// Story represents a development story
type Story struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Author      string    `json:"author"`
	Tags        []string  `json:"tags"`
	JiraKey     string    `json:"jira_key,omitempty"`
	Number      int       `json:"number,omitempty"`
	Commits     []Commit  `json:"commits,omitempty"`
	Files       []File    `json:"files,omitempty"`
	Filename    string    `json:"-"`
}

// NewStory creates a new story with the given title and description
func NewStory(title, description, author string) (*Story, error) {
	now := time.Now()
	story := &Story{
		ID:          utils.GenerateID(),
		Title:       title,
		Description: description,
		Status:      "open",
		CreatedAt:   now,
		UpdatedAt:   now,
		Author:      author,
		Tags:        []string{},
		Number:      0,
	}

	// Get project name from git config
	projectName, _ := utils.GitClient.GetConfig("current.project")

	// Try to create a git branch for the story
	branchName := utils.GenerateBranchName(title, story.ID, story.Number, projectName)
	if err := utils.CreateBranch(branchName); err != nil {
		// If we can't create a branch, it's not a critical error
		// The story will still be created, but we'll log the error
		fmt.Printf("Warning: Failed to create git branch: %v\n", err)
	}

	return story, nil
}

// NewStoryWithNumber creates a new story with the given title, description, and number
func NewStoryWithNumber(title, description, author string, number int) (*Story, error) {
	if number <= 0 {
		return nil, fmt.Errorf("number must be greater than 0")
	}

	if title == "" {
		return nil, fmt.Errorf("title is required when creating a story with a number")
	}

	now := time.Now()
	story := &Story{
		ID:          utils.GenerateID(),
		Title:       title,
		Description: description,
		Status:      "open",
		CreatedAt:   now,
		UpdatedAt:   now,
		Author:      author,
		Tags:        []string{},
		Number:      number,
	}

	// Get project name from git config
	projectName, _ := utils.GitClient.GetConfig("current.project")

	// Try to create a git branch for the story
	branchName := utils.GenerateBranchName(title, story.ID, story.Number, projectName)
	if err := utils.CreateBranch(branchName); err != nil {
		// If we can't create a branch, it's not a critical error
		// The story will still be created, but we'll log the error
		fmt.Printf("Warning: Failed to create git branch: %v\n", err)
	}

	return story, nil
}

// Save saves the story to disk
func (s *Story) Save() error {
	if s.Filename == "" {
		s.Filename = fmt.Sprintf("%s.yaml", s.ID)
	}

	return SaveStory(s)
}

// LoadStory loads a story from a file
func LoadStory(filename string) (*Story, error) {
	storiesDir, err := GetStoriesDir()
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(storiesDir, filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read story file: %w", err)
	}

	var story Story
	if err := yaml.Unmarshal(data, &story); err != nil {
		return nil, fmt.Errorf("failed to unmarshal story: %w", err)
	}

	return &story, nil
}

// ListStories returns a list of all stories
func ListStories() ([]*Story, error) {
	storiesDir, err := GetStoriesDir()
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(storiesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read stories directory: %w", err)
	}

	var stories []*Story
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
			story, err := LoadStory(file.Name())
			if err != nil {
				return nil, fmt.Errorf("failed to load story %s: %w", file.Name(), err)
			}
			stories = append(stories, story)
		}
	}

	// Sort stories by creation date in descending order (newest first)
	sort.Slice(stories, func(i, j int) bool {
		return stories[i].CreatedAt.After(stories[j].CreatedAt)
	})

	return stories, nil
}

// SaveStory saves a story to a file
func SaveStory(story *Story) error {
	storiesDir, err := GetStoriesDir()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(story)
	if err != nil {
		return fmt.Errorf("failed to marshal story: %w", err)
	}

	filePath := filepath.Join(storiesDir, story.Filename)
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write story file: %w", err)
	}

	return nil
}

// AddCommit adds a commit to the story
func (s *Story) AddCommit(hash, message, author string, timestamp time.Time) {
	s.Commits = append(s.Commits, Commit{
		Hash:      hash,
		Message:   message,
		Author:    author,
		Timestamp: timestamp,
	})
	s.UpdatedAt = time.Now()
}

// AddFile adds a file to the story
func (s *Story) AddFile(path, status string) {
	s.Files = append(s.Files, File{
		Path:      path,
		Status:    status,
		Timestamp: time.Now(),
	})
	s.UpdatedAt = time.Now()
}

// GetCommits returns all commits associated with the story
func (s *Story) GetCommits() []Commit {
	return s.Commits
}

// GetFiles returns all files associated with the story
func (s *Story) GetFiles() []File {
	return s.Files
}

// GetStoriesDir returns the directory where stories are stored
func GetStoriesDir() (string, error) {
	// Try to get repository-specific stories directory first
	repoConfigDir, err := utils.GetRepoConfigDir()
	if err == nil {
		storiesDir := filepath.Join(repoConfigDir, "stories")
		if err := utils.EnsureDir(storiesDir); err != nil {
			return "", fmt.Errorf("failed to create stories directory: %w", err)
		}
		return storiesDir, nil
	}

	// If no repository is found, use global stories directory
	globalConfigDir, err := utils.GetConfigDir()
	if err != nil {
		return "", err
	}

	storiesDir := filepath.Join(globalConfigDir, "stories")
	if err := utils.EnsureDir(storiesDir); err != nil {
		return "", fmt.Errorf("failed to create stories directory: %w", err)
	}

	return storiesDir, nil
}
