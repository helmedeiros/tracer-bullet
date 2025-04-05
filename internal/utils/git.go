package utils

import (
	"strings"
)

// GitOperations defines the interface for git operations
type GitOperations interface {
	Init() error
	SetConfig(key, value string) error
	GetConfig(key string) (string, error)
	ParseRevision(rev string) (string, error)
	Commit(message string) error
	GetCurrentHead() (string, error)
	GetAuthor() (string, error)
	GetChangedFiles() ([]string, error)
	GetGitRoot() (string, error)
}

// RealGit implements GitOperations using actual git commands
type RealGit struct{}

// NewRealGit creates a new RealGit instance
func NewRealGit() GitOperations {
	return &RealGit{}
}

// MockGit implements GitOperations for testing
type MockGit struct {
	InitFunc            func() error
	SetConfigFunc       func(key, value string) error
	GetConfigFunc       func(key string) (string, error)
	ParseRevisionFunc   func(rev string) (string, error)
	CommitFunc          func(message string) error
	GetCurrentHeadFunc  func() (string, error)
	GetAuthorFunc       func() (string, error)
	GetChangedFilesFunc func() ([]string, error)
	GetGitRootFunc      func() (string, error)
}

// NewMockGit creates a new MockGit instance
func NewMockGit() GitOperations {
	return &MockGit{
		InitFunc: func() error {
			return nil
		},
		SetConfigFunc: func(key, value string) error {
			return nil
		},
		GetConfigFunc: func(key string) (string, error) {
			return "", nil
		},
		ParseRevisionFunc: func(rev string) (string, error) {
			return "", nil
		},
		CommitFunc: func(message string) error {
			return nil
		},
		GetCurrentHeadFunc: func() (string, error) {
			return "", nil
		},
		GetAuthorFunc: func() (string, error) {
			return "", nil
		},
		GetChangedFilesFunc: func() ([]string, error) {
			return nil, nil
		},
		GetGitRootFunc: func() (string, error) {
			return "", nil
		},
	}
}

// Init initializes a git repository
func (g *RealGit) Init() error {
	_, err := RunCommand("git", "init")
	return err
}

// SetConfig sets a git configuration value
func (g *RealGit) SetConfig(key, value string) error {
	_, err := RunCommand("git", "config", key, value)
	return err
}

// GetConfig gets a git configuration value
func (g *RealGit) GetConfig(key string) (string, error) {
	return RunCommand("git", "config", key)
}

// ParseRevision parses a git revision
func (g *RealGit) ParseRevision(rev string) (string, error) {
	return RunCommand("git", "rev-parse", rev)
}

// Commit creates a git commit
func (g *RealGit) Commit(message string) error {
	_, err := RunCommand("git", "commit", "-m", message)
	return err
}

// GetCurrentHead gets the current git head
func (g *RealGit) GetCurrentHead() (string, error) {
	return RunCommand("git", "rev-parse", "HEAD")
}

// GetAuthor gets the git author
func (g *RealGit) GetAuthor() (string, error) {
	return RunCommand("git", "config", "user.name")
}

// GetChangedFiles gets the list of changed files
func (g *RealGit) GetChangedFiles() ([]string, error) {
	output, err := RunCommand("git", "diff", "--name-only")
	if err != nil {
		return nil, err
	}
	return splitLines(output), nil
}

// GetGitRoot gets the root directory of the git repository
func (g *RealGit) GetGitRoot() (string, error) {
	return RunCommand("git", "rev-parse", "--show-toplevel")
}

// Init initializes a git repository (mock implementation)
func (g *MockGit) Init() error {
	if g.InitFunc != nil {
		return g.InitFunc()
	}
	return nil
}

// SetConfig sets a git configuration value (mock implementation)
func (g *MockGit) SetConfig(key, value string) error {
	if g.SetConfigFunc != nil {
		return g.SetConfigFunc(key, value)
	}
	return nil
}

// GetConfig gets a git configuration value (mock implementation)
func (g *MockGit) GetConfig(key string) (string, error) {
	if g.GetConfigFunc != nil {
		return g.GetConfigFunc(key)
	}
	return "", nil
}

// ParseRevision parses a git revision (mock implementation)
func (g *MockGit) ParseRevision(rev string) (string, error) {
	if g.ParseRevisionFunc != nil {
		return g.ParseRevisionFunc(rev)
	}
	return "", nil
}

// Commit creates a git commit (mock implementation)
func (g *MockGit) Commit(message string) error {
	if g.CommitFunc != nil {
		return g.CommitFunc(message)
	}
	return nil
}

// GetCurrentHead gets the current git head (mock implementation)
func (g *MockGit) GetCurrentHead() (string, error) {
	if g.GetCurrentHeadFunc != nil {
		return g.GetCurrentHeadFunc()
	}
	return "", nil
}

// GetAuthor gets the git author (mock implementation)
func (g *MockGit) GetAuthor() (string, error) {
	if g.GetAuthorFunc != nil {
		return g.GetAuthorFunc()
	}
	return "", nil
}

// GetChangedFiles gets the list of changed files (mock implementation)
func (g *MockGit) GetChangedFiles() ([]string, error) {
	if g.GetChangedFilesFunc != nil {
		return g.GetChangedFilesFunc()
	}
	return nil, nil
}

// GetGitRoot gets the root directory of the git repository (mock implementation)
func (g *MockGit) GetGitRoot() (string, error) {
	if g.GetGitRootFunc != nil {
		return g.GetGitRootFunc()
	}
	return "", nil
}

// splitLines splits a string into lines and trims whitespace
func splitLines(s string) []string {
	lines := strings.Split(s, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
