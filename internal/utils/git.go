package utils

import (
	"fmt"
	"os"
	"os/exec"
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
	CreateBranch(branchName string) error
	SwitchBranch(branchName string) error
	BranchExists(branchName string) (bool, error)
}

// RealGit implements GitOperations using actual git commands
type RealGit struct{}

// NewRealGit creates a new RealGit instance
func NewRealGit() GitOperations {
	return &RealGit{}
}

// BaseMockGit provides default implementations for all GitOperations methods
type BaseMockGit struct {
	InitFunc            func() error
	SetConfigFunc       func(key, value string) error
	GetConfigFunc       func(key string) (string, error)
	ParseRevisionFunc   func(rev string) (string, error)
	CommitFunc          func(message string) error
	GetCurrentHeadFunc  func() (string, error)
	GetAuthorFunc       func() (string, error)
	GetChangedFilesFunc func() ([]string, error)
	GetGitRootFunc      func() (string, error)
	CreateBranchFunc    func(branchName string) error
	SwitchBranchFunc    func(branchName string) error
	BranchExistsFunc    func(branchName string) (bool, error)
}

// NewBaseMockGit creates a new BaseMockGit with default implementations
func NewBaseMockGit() *BaseMockGit {
	return &BaseMockGit{
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
		CreateBranchFunc: func(branchName string) error {
			return nil
		},
		SwitchBranchFunc: func(branchName string) error {
			return nil
		},
		BranchExistsFunc: func(branchName string) (bool, error) {
			return false, nil
		},
	}
}

// MockGit implements GitOperations for testing
type MockGit struct {
	*BaseMockGit
}

// NewMockGit creates a new MockGit instance
func NewMockGit() GitOperations {
	return &MockGit{
		BaseMockGit: NewBaseMockGit(),
	}
}

// Init checks if we're in a git repository
func (g *RealGit) Init() error {
	// Check if we're in a git repository
	if _, err := RunCommand("git", "rev-parse", "--is-inside-work-tree"); err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}
	return nil
}

// SetConfig sets a git configuration value
func (g *RealGit) SetConfig(key, value string) error {
	// Try to get the git root directory
	gitRoot, err := g.GetGitRoot()
	if err != nil {
		// If we can't get the git root, use the home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		gitRoot = home
	}

	// Set the config value using the git root directory
	_, err = RunCommand("git", "-C", gitRoot, "config", key, value)
	if err != nil {
		return fmt.Errorf("failed to set git config %s: %w", key, err)
	}
	return nil
}

// GetConfig gets a git configuration value
func (g *RealGit) GetConfig(key string) (string, error) {
	// Try to get the git root directory
	gitRoot, err := g.GetGitRoot()
	if err != nil {
		// If we can't get the git root, use the home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		gitRoot = home
	}

	// Get the config value using the git root directory
	value, err := RunCommand("git", "-C", gitRoot, "config", key)
	if err != nil {
		// If the config doesn't exist, return empty string
		if strings.Contains(err.Error(), "exit status 1") {
			return "", nil
		}
		return "", fmt.Errorf("failed to get git config %s: %w", key, err)
	}
	return strings.TrimSpace(value), nil
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
	// Try to get the git root directory
	output, err := RunCommand("git", "rev-parse", "--show-toplevel")
	if err != nil {
		// If we're not in a git repository, try to find the nearest parent git repository
		output, err = RunCommand("git", "rev-parse", "--git-dir")
		if err != nil {
			return "", fmt.Errorf("not in a git repository")
		}
	}
	return strings.TrimSpace(output), nil
}

// CreateBranch creates a new git branch
func (g *RealGit) CreateBranch(branchName string) error {
	_, err := RunCommand("git", "checkout", "-b", branchName)
	return err
}

// SwitchBranch switches to an existing git branch
func (g *RealGit) SwitchBranch(branchName string) error {
	_, err := RunCommand("git", "checkout", branchName)
	return err
}

// BranchExists checks if a branch exists
func (g *RealGit) BranchExists(branchName string) (bool, error) {
	_, err := RunCommand("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branchName)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Init initializes a git repository (mock implementation)
func (g *MockGit) Init() error {
	return g.InitFunc()
}

// SetConfig sets a git configuration value (mock implementation)
func (g *MockGit) SetConfig(key, value string) error {
	return g.SetConfigFunc(key, value)
}

// GetConfig gets a git configuration value (mock implementation)
func (g *MockGit) GetConfig(key string) (string, error) {
	return g.GetConfigFunc(key)
}

// ParseRevision parses a git revision (mock implementation)
func (g *MockGit) ParseRevision(rev string) (string, error) {
	return g.ParseRevisionFunc(rev)
}

// Commit creates a git commit (mock implementation)
func (g *MockGit) Commit(message string) error {
	return g.CommitFunc(message)
}

// GetCurrentHead gets the current git head (mock implementation)
func (g *MockGit) GetCurrentHead() (string, error) {
	return g.GetCurrentHeadFunc()
}

// GetAuthor gets the git author (mock implementation)
func (g *MockGit) GetAuthor() (string, error) {
	return g.GetAuthorFunc()
}

// GetChangedFiles gets the list of changed files (mock implementation)
func (g *MockGit) GetChangedFiles() ([]string, error) {
	return g.GetChangedFilesFunc()
}

// GetGitRoot gets the root directory of the git repository (mock implementation)
func (g *MockGit) GetGitRoot() (string, error) {
	return g.GetGitRootFunc()
}

// CreateBranch creates a new git branch (mock implementation)
func (g *MockGit) CreateBranch(branchName string) error {
	return g.CreateBranchFunc(branchName)
}

// SwitchBranch switches to an existing git branch (mock implementation)
func (g *MockGit) SwitchBranch(branchName string) error {
	return g.SwitchBranchFunc(branchName)
}

// BranchExists checks if a branch exists (mock implementation)
func (g *MockGit) BranchExists(branchName string) (bool, error) {
	return g.BranchExistsFunc(branchName)
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
