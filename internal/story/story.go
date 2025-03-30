package story

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/helmedeiros/tracer-bullet/internal/config"
	"github.com/helmedeiros/tracer-bullet/internal/utils"
)

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
	// Get config to determine story directory
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create story directory if it doesn't exist
	storyDir := filepath.Join(cfg.StoryDir)
	if err := os.MkdirAll(storyDir, utils.DefaultDirPerm); err != nil {
		return fmt.Errorf("failed to create story directory: %w", err)
	}

	// Create story file
	storyFile := filepath.Join(storyDir, fmt.Sprintf("%s%s", s.ID, config.DefaultStoryExt))
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal story: %w", err)
	}

	if err := os.WriteFile(storyFile, data, utils.DefaultFilePerm); err != nil {
		return fmt.Errorf("failed to write story file: %w", err)
	}

	return nil
}

// LoadStory loads a story by ID
func LoadStory(id string) (*Story, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	storyFile := filepath.Join(cfg.StoryDir, fmt.Sprintf("%s%s", id, config.DefaultStoryExt))
	data, err := os.ReadFile(storyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read story file: %w", err)
	}

	var story Story
	if err := json.Unmarshal(data, &story); err != nil {
		return nil, fmt.Errorf("failed to unmarshal story: %w", err)
	}

	return &story, nil
}

// ListStories returns all stories
func ListStories() ([]*Story, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	storyDir := cfg.StoryDir
	if _, err := os.Stat(storyDir); os.IsNotExist(err) {
		return []*Story{}, nil
	}

	files, err := os.ReadDir(storyDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read story directory: %w", err)
	}

	var stories []*Story
	for _, file := range files {
		if filepath.Ext(file.Name()) != config.DefaultStoryExt {
			continue
		}

		id := file.Name()[:len(file.Name())-len(config.DefaultStoryExt)]
		story, err := LoadStory(id)
		if err != nil {
			return nil, fmt.Errorf("failed to load story %s: %w", id, err)
		}

		stories = append(stories, story)
	}

	return stories, nil
}
