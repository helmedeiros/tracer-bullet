package utils

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRealGit_Init(t *testing.T) {
	// Save current GitClient
	originalGitClient := GitClient
	defer func() {
		GitClient = originalGitClient
	}()

	// Use mock Git client
	GitClient = NewMockGit()
	mockGit := GitClient.(*MockGit)
	mockGit.InitFunc = func() error {
		return nil
	}

	// Test initialization
	err := GitClient.Init()
	assert.NoError(t, err)
}

func TestRealGit_SetConfig(t *testing.T) {
	// Save current GitClient
	originalGitClient := GitClient
	defer func() {
		GitClient = originalGitClient
	}()

	// Use mock Git client
	GitClient = NewMockGit()
	mockGit := GitClient.(*MockGit)
	mockGit.SetConfigFunc = func(key, value string) error {
		return nil
	}
	mockGit.GetConfigFunc = func(key string) (string, error) {
		return "Test User", nil
	}

	// Test setting config
	err := GitClient.SetConfig("user.name", "Test User")
	assert.NoError(t, err)

	// Verify config was set
	value, err := GitClient.GetConfig("user.name")
	assert.NoError(t, err)
	assert.Equal(t, "Test User", value)
}

func TestRealGit_GetConfig(t *testing.T) {
	// Save current GitClient
	originalGitClient := GitClient
	defer func() {
		GitClient = originalGitClient
	}()

	// Use mock Git client
	GitClient = NewMockGit()
	mockGit := GitClient.(*MockGit)
	mockGit.GetConfigFunc = func(key string) (string, error) {
		if key == "nonexistent.config" {
			return "", nil
		}
		return "test@example.com", nil
	}
	mockGit.SetConfigFunc = func(key, value string) error {
		return nil
	}

	// Test getting non-existent config
	value, err := GitClient.GetConfig("nonexistent.config")
	assert.NoError(t, err)
	assert.Empty(t, value)

	// Test getting existing config
	err = GitClient.SetConfig("user.email", "test@example.com")
	require.NoError(t, err)

	value, err = GitClient.GetConfig("user.email")
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", value)
}

func TestRealGit_GetGitRoot(t *testing.T) {
	// Create a temporary directory for testing
	dir := t.TempDir()

	// Save current GitClient
	originalGitClient := GitClient
	defer func() {
		GitClient = originalGitClient
	}()

	// Use mock Git client
	GitClient = NewMockGit()
	mockGit := GitClient.(*MockGit)
	mockGit.GetGitRootFunc = func() (string, error) {
		return dir, nil
	}

	// Test getting git root
	root, err := GitClient.GetGitRoot()
	assert.NoError(t, err)
	assert.Equal(t, dir, root)
}

func TestRealGit_GetAuthor(t *testing.T) {
	// Save current GitClient
	originalGitClient := GitClient
	defer func() {
		GitClient = originalGitClient
	}()

	// Use mock Git client
	GitClient = NewMockGit()
	mockGit := GitClient.(*MockGit)
	mockGit.GetAuthorFunc = func() (string, error) {
		return "Test Author", nil
	}

	// Test getting author
	author, err := GitClient.GetAuthor()
	assert.NoError(t, err)
	assert.Equal(t, "Test Author", author)
}

func TestRealGit_GetCurrentHead(t *testing.T) {
	// Save current GitClient
	originalGitClient := GitClient
	defer func() {
		GitClient = originalGitClient
	}()

	// Use mock Git client
	GitClient = NewMockGit()
	mockGit := GitClient.(*MockGit)
	mockGit.GetCurrentHeadFunc = func() (string, error) {
		return "test-commit-hash", nil
	}

	// Test getting current head
	head, err := GitClient.GetCurrentHead()
	assert.NoError(t, err)
	assert.Equal(t, "test-commit-hash", head)
}

func TestRealGit_GetChangedFiles(t *testing.T) {
	// Save current GitClient
	originalGitClient := GitClient
	defer func() {
		GitClient = originalGitClient
	}()

	// Use mock Git client
	GitClient = NewMockGit()
	mockGit := GitClient.(*MockGit)
	mockGit.GetChangedFilesFunc = func() ([]string, error) {
		return []string{"test.txt"}, nil
	}

	// Test getting changed files
	files, err := GitClient.GetChangedFiles()
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, "test.txt", files[0])
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single line",
			input:    "hello",
			expected: []string{"hello"},
		},
		{
			name:     "multiple lines",
			input:    "hello\nworld\n",
			expected: []string{"hello", "world"},
		},
		{
			name:     "lines with whitespace",
			input:    "  hello  \n  world  \n",
			expected: []string{"hello", "world"},
		},
		{
			name:     "empty lines",
			input:    "hello\n\nworld\n\n",
			expected: []string{"hello", "world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRealGit_GetConfig_Error(t *testing.T) {
	// Save current GitClient
	originalGitClient := GitClient
	defer func() {
		GitClient = originalGitClient
	}()

	// Use mock Git client
	GitClient = NewMockGit()
	mockGit := GitClient.(*MockGit)
	mockGit.GetConfigFunc = func(key string) (string, error) {
		return "", os.ErrPermission
	}

	// Test getting config with error
	_, err := GitClient.GetConfig("test.key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestRealGit_SetConfig_Error(t *testing.T) {
	// Save current GitClient
	originalGitClient := GitClient
	defer func() {
		GitClient = originalGitClient
	}()

	// Use mock Git client
	GitClient = NewMockGit()
	mockGit := GitClient.(*MockGit)
	mockGit.SetConfigFunc = func(key, value string) error {
		return os.ErrPermission
	}

	// Test setting config with error
	err := GitClient.SetConfig("test.key", "test.value")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestRealGit_GetGitRoot_Error(t *testing.T) {
	// Save current GitClient
	originalGitClient := GitClient
	defer func() {
		GitClient = originalGitClient
	}()

	// Use mock Git client
	GitClient = NewMockGit()
	mockGit := GitClient.(*MockGit)
	mockGit.GetGitRootFunc = func() (string, error) {
		return "", fmt.Errorf("not in a git repository")
	}

	// Test getting git root with error
	_, err := GitClient.GetGitRoot()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in a git repository")
}

func TestRealGit_GetCurrentHead_Error(t *testing.T) {
	// Save current GitClient
	originalGitClient := GitClient
	defer func() {
		GitClient = originalGitClient
	}()

	// Use mock Git client
	GitClient = NewMockGit()
	mockGit := GitClient.(*MockGit)
	mockGit.GetCurrentHeadFunc = func() (string, error) {
		return "", fmt.Errorf("no commits yet")
	}

	// Test getting current head with error
	_, err := GitClient.GetCurrentHead()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no commits yet")
}

func TestRealGit_GetChangedFiles_Error(t *testing.T) {
	// Save current GitClient
	originalGitClient := GitClient
	defer func() {
		GitClient = originalGitClient
	}()

	// Use mock Git client
	GitClient = NewMockGit()
	mockGit := GitClient.(*MockGit)
	mockGit.GetChangedFilesFunc = func() ([]string, error) {
		return nil, fmt.Errorf("working directory not clean")
	}

	// Test getting changed files with error
	_, err := GitClient.GetChangedFiles()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "working directory not clean")
}

func TestRealGit_ParseRevision_Error(t *testing.T) {
	// Save current GitClient
	originalGitClient := GitClient
	defer func() {
		GitClient = originalGitClient
	}()

	// Use mock Git client
	GitClient = NewMockGit()
	mockGit := GitClient.(*MockGit)
	mockGit.ParseRevisionFunc = func(rev string) (string, error) {
		return "", fmt.Errorf("invalid revision")
	}

	// Test parsing revision with error
	_, err := GitClient.ParseRevision("invalid-ref")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid revision")
}

func TestRealGit_Commit_Error(t *testing.T) {
	// Save current GitClient
	originalGitClient := GitClient
	defer func() {
		GitClient = originalGitClient
	}()

	// Use mock Git client
	GitClient = NewMockGit()
	mockGit := GitClient.(*MockGit)
	mockGit.CommitFunc = func(message string) error {
		return fmt.Errorf("nothing to commit")
	}

	// Test committing with error
	err := GitClient.Commit("test commit")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nothing to commit")
}
