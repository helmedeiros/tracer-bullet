package story

import (
	"fmt"
	"os"
	"path/filepath"
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
	Commits     []Commit  `json:"commits,omitempty"`
	Files       []File    `json:"files,omitempty"`
	Filename    string    `json:"-"`
}

// NewStory creates a new story with the given title and description
func NewStory(title, description, author string) (*Story, error) {
	if title == "" {
		return nil, fmt.Errorf("title cannot be empty")
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
