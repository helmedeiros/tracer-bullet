package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCommand(t *testing.T) {
	tests := []struct {
		name       string
		command    string
		args       []string
		wantErr    bool
		wantOutput string
		errMsg     string
	}{
		{
			name:       "echo command",
			command:    "echo",
			args:       []string{"hello world"},
			wantErr:    false,
			wantOutput: "hello world\n",
		},
		{
			name:    "invalid command",
			command: "nonexistentcommand",
			args:    []string{},
			wantErr: true,
			errMsg:  "executable file not found",
		},
		{
			name:    "command with invalid args",
			command: "ls",
			args:    []string{"--invalid-flag"},
			wantErr: true,
			errMsg:  "exit status",
		},
		{
			name:       "command with multiple args",
			command:    "echo",
			args:       []string{"hello", "world"},
			wantErr:    false,
			wantOutput: "hello world\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := RunCommand(tt.command, tt.args...)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOutput, output)
		})
	}
}

func TestGenerateID(t *testing.T) {
	// Test multiple generations to ensure uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := GenerateID()
		assert.NotEmpty(t, id)
		assert.False(t, ids[id], "duplicate ID generated")
		ids[id] = true
	}
}

func TestGetConfigDir(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		teardown func(t *testing.T, path string)
	}{
		{
			name: "default config dir",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				TestConfigDir = dir
				return dir
			},
			teardown: func(t *testing.T, path string) {
				TestConfigDir = ""
			},
		},
		{
			name: "custom config dir",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				TestConfigDir = filepath.Join(dir, "custom")
				err := os.MkdirAll(TestConfigDir, 0755)
				require.NoError(t, err)
				return dir
			},
			teardown: func(t *testing.T, path string) {
				TestConfigDir = ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			defer tt.teardown(t, dir)

			configDir, err := GetConfigDir()
			assert.NoError(t, err)
			assert.NotEmpty(t, configDir)
			assert.DirExists(t, configDir)
		})
	}
}

func TestCheckDirWritable(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		teardown func(t *testing.T, path string)
		wantErr  bool
		errMsg   string
	}{
		{
			name: "writable directory",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			teardown: func(t *testing.T, path string) {},
			wantErr:  false,
		},
		{
			name: "non-existent directory",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				return filepath.Join(dir, "nonexistent")
			},
			teardown: func(t *testing.T, path string) {},
			wantErr:  false, // Should create the directory
		},
		{
			name: "unwritable directory",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				err := os.Chmod(dir, 0555)
				require.NoError(t, err)
				return dir
			},
			teardown: func(t *testing.T, path string) {
				err := os.Chmod(path, 0755)
				assert.NoError(t, err)
			},
			wantErr: true,
			errMsg:  "permission denied",
		},
		{
			name: "path is a file",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				file := filepath.Join(dir, "file")
				err := os.WriteFile(file, []byte("test"), 0644)
				require.NoError(t, err)
				return file
			},
			teardown: func(t *testing.T, path string) {},
			wantErr:  true,
			errMsg:   "not a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			defer tt.teardown(t, path)

			err := checkDirWritable(path)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestEnsureDir(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		teardown func(t *testing.T, path string)
		wantErr  bool
		errMsg   string
	}{
		{
			name: "create new directory",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				return filepath.Join(dir, "newdir")
			},
			teardown: func(t *testing.T, path string) {},
			wantErr:  false,
		},
		{
			name: "existing directory",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			teardown: func(t *testing.T, path string) {},
			wantErr:  false,
		},
		{
			name: "parent directory not writable",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				err := os.Chmod(dir, 0555)
				require.NoError(t, err)
				return filepath.Join(dir, "newdir")
			},
			teardown: func(t *testing.T, path string) {
				parent := filepath.Dir(path)
				err := os.Chmod(parent, 0755)
				assert.NoError(t, err)
			},
			wantErr: true,
			errMsg:  "permission denied",
		},
		{
			name: "path is a file",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				file := filepath.Join(dir, "file")
				err := os.WriteFile(file, []byte("test"), 0644)
				require.NoError(t, err)
				return file
			},
			teardown: func(t *testing.T, path string) {},
			wantErr:  false, // EnsureDir only creates if it doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			defer tt.teardown(t, path)

			err := EnsureDir(path)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGetGitRoot(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		teardown func(t *testing.T, path string)
		wantErr  bool
	}{
		{
			name: "git repository",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				GitClient = NewMockGit()
				GitClient.(*MockGit).GetGitRootFunc = func() (string, error) {
					return dir, nil
				}
				return dir
			},
			teardown: func(t *testing.T, path string) {
				GitClient = NewRealGit()
			},
			wantErr: false,
		},
		{
			name: "not a git repository",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				GitClient = NewMockGit()
				GitClient.(*MockGit).GetGitRootFunc = func() (string, error) {
					return "", os.ErrNotExist
				}
				return dir
			},
			teardown: func(t *testing.T, path string) {
				GitClient = NewRealGit()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			defer tt.teardown(t, dir)

			root, err := GetGitRoot()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, dir, root)
		})
	}
}

func TestRunGitInit(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		teardown func(t *testing.T, path string)
		wantErr  bool
	}{
		{
			name: "successful git init",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				GitClient = NewMockGit()
				GitClient.(*MockGit).InitFunc = func() error {
					return nil
				}
				return dir
			},
			teardown: func(t *testing.T, path string) {
				GitClient = NewRealGit()
			},
			wantErr: false,
		},
		{
			name: "failed git init",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				GitClient = NewMockGit()
				GitClient.(*MockGit).InitFunc = func() error {
					return os.ErrPermission
				}
				return dir
			},
			teardown: func(t *testing.T, path string) {
				GitClient = NewRealGit()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			defer tt.teardown(t, dir)

			err := RunGitInit()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestRunGitConfig(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		setup    func(t *testing.T) string
		teardown func(t *testing.T, path string)
		wantErr  bool
	}{
		{
			name:  "successful git config",
			key:   "user.name",
			value: "Test User",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				GitClient = NewMockGit()
				GitClient.(*MockGit).SetConfigFunc = func(key, value string) error {
					return nil
				}
				return dir
			},
			teardown: func(t *testing.T, path string) {
				GitClient = NewRealGit()
			},
			wantErr: false,
		},
		{
			name:  "failed git config",
			key:   "user.name",
			value: "Test User",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				GitClient = NewMockGit()
				GitClient.(*MockGit).SetConfigFunc = func(key, value string) error {
					return os.ErrPermission
				}
				return dir
			},
			teardown: func(t *testing.T, path string) {
				GitClient = NewRealGit()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			defer tt.teardown(t, dir)

			err := RunGitConfig(tt.key, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestFileExists(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		teardown func(t *testing.T, path string)
		exists   bool
	}{
		{
			name: "file exists",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				file := filepath.Join(dir, "test.txt")
				err := os.WriteFile(file, []byte("test"), 0644)
				require.NoError(t, err)
				return file
			},
			teardown: func(t *testing.T, path string) {},
			exists:   true,
		},
		{
			name: "file does not exist",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				return filepath.Join(dir, "nonexistent.txt")
			},
			teardown: func(t *testing.T, path string) {},
			exists:   false,
		},
		{
			name: "directory exists",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				return dir
			},
			teardown: func(t *testing.T, path string) {},
			exists:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			defer tt.teardown(t, path)

			exists := FileExists(path)
			assert.Equal(t, tt.exists, exists)
		})
	}
}

func TestGetHomeDir(t *testing.T) {
	// Save original home directory
	originalHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("HOME", originalHome)
	}()

	tests := []struct {
		name     string
		setup    func(t *testing.T)
		teardown func(t *testing.T)
		wantErr  bool
	}{
		{
			name: "valid home directory",
			setup: func(t *testing.T) {
				dir := t.TempDir()
				os.Setenv("HOME", dir)
			},
			teardown: func(t *testing.T) {},
			wantErr:  false,
		},
		{
			name: "empty home directory",
			setup: func(t *testing.T) {
				os.Setenv("HOME", "")
			},
			teardown: func(t *testing.T) {},
			wantErr:  true,
		},
		{
			name: "unset home directory",
			setup: func(t *testing.T) {
				os.Unsetenv("HOME")
			},
			teardown: func(t *testing.T) {},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)
			defer tt.teardown(t)

			home, err := GetHomeDir()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, home)
			assert.DirExists(t, home)
		})
	}
}

func TestGetRepoConfigDir(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		teardown func(t *testing.T, path string)
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid git repository",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				GitClient = NewMockGit()
				GitClient.(*MockGit).GetGitRootFunc = func() (string, error) {
					return dir, nil
				}
				return dir
			},
			teardown: func(t *testing.T, path string) {
				GitClient = NewRealGit()
			},
			wantErr: false,
		},
		{
			name: "not a git repository",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				GitClient = NewMockGit()
				GitClient.(*MockGit).GetGitRootFunc = func() (string, error) {
					return "", os.ErrNotExist
				}
				return dir
			},
			teardown: func(t *testing.T, path string) {
				GitClient = NewRealGit()
			},
			wantErr: true,
			errMsg:  "file does not exist",
		},
		{
			name: "unwritable directory",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				GitClient = NewMockGit()
				GitClient.(*MockGit).GetGitRootFunc = func() (string, error) {
					return dir, nil
				}
				// Make the directory read-only
				err := os.Chmod(dir, 0555)
				require.NoError(t, err)
				return dir
			},
			teardown: func(t *testing.T, path string) {
				GitClient = NewRealGit()
				// Restore permissions
				err := os.Chmod(path, 0755)
				assert.NoError(t, err)
			},
			wantErr: true,
			errMsg:  "permission denied",
		},
		{
			name: "git root is a file",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				file := filepath.Join(dir, "file")
				err := os.WriteFile(file, []byte("test"), 0644)
				require.NoError(t, err)
				GitClient = NewMockGit()
				GitClient.(*MockGit).GetGitRootFunc = func() (string, error) {
					return file, nil
				}
				return dir
			},
			teardown: func(t *testing.T, path string) {
				GitClient = NewRealGit()
			},
			wantErr: true,
			errMsg:  "not a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			defer tt.teardown(t, dir)

			configDir, err := GetRepoConfigDir()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, configDir)
			assert.DirExists(t, configDir)
		})
	}
}

func TestParseRevision(t *testing.T) {
	// Save current GitClient
	originalGitClient := GitClient
	defer func() {
		GitClient = originalGitClient
	}()

	tests := []struct {
		name    string
		rev     string
		result  string
		wantErr bool
		errMsg  string
	}{
		{
			name:   "valid revision",
			rev:    "HEAD",
			result: "abcdef1234567890",
		},
		{
			name:    "invalid revision",
			rev:     "invalid-ref",
			wantErr: true,
			errMsg:  "invalid revision",
		},
		{
			name:    "empty revision",
			rev:     "",
			wantErr: true,
			errMsg:  "empty revision",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use mock Git client
			GitClient = NewMockGit()
			mockGit := GitClient.(*MockGit)
			mockGit.ParseRevisionFunc = func(rev string) (string, error) {
				if rev == "HEAD" {
					return tt.result, nil
				}
				return "", fmt.Errorf("%s", tt.errMsg)
			}

			// Test parsing revision
			value, err := GitClient.ParseRevision(tt.rev)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.result, value)
		})
	}
}

func TestGenerateBranchName(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		id       string
		expected string
	}{
		{
			name:     "simple title",
			title:    "Test Story",
			id:       "story-123",
			expected: "test-story",
		},
		{
			name:     "title with special characters",
			title:    "Test Story: Fix Bug #123",
			id:       "story-123",
			expected: "test-story-fix-bug-123",
		},
		{
			name:     "title with multiple spaces",
			title:    "Test  Story  With  Spaces",
			id:       "story-123",
			expected: "test-story-with-spaces",
		},
		{
			name:     "empty title uses ID",
			title:    "",
			id:       "story-123",
			expected: "story-123",
		},
		{
			name:     "title with underscores",
			title:    "test_story_with_underscores",
			id:       "story-123",
			expected: "test-story-with-underscores",
		},
		{
			name:     "title with dots",
			title:    "test.story.with.dots",
			id:       "story-123",
			expected: "test-story-with-dots",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateBranchName(tt.title, tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateBranch(t *testing.T) {
	tests := []struct {
		name          string
		branchName    string
		branchExists  bool
		createError   error
		switchError   error
		existsError   error
		expectedError bool
	}{
		{
			name:          "create new branch",
			branchName:    "test-branch",
			branchExists:  false,
			expectedError: false,
		},
		{
			name:          "switch to existing branch",
			branchName:    "test-branch",
			branchExists:  true,
			expectedError: false,
		},
		{
			name:          "error checking branch existence",
			branchName:    "test-branch",
			existsError:   fmt.Errorf("git error"),
			expectedError: true,
		},
		{
			name:          "error creating branch",
			branchName:    "test-branch",
			branchExists:  false,
			createError:   fmt.Errorf("git error"),
			expectedError: true,
		},
		{
			name:          "error switching branch",
			branchName:    "test-branch",
			branchExists:  true,
			switchError:   fmt.Errorf("git error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save current GitClient
			originalGitClient := GitClient
			defer func() {
				GitClient = originalGitClient
			}()

			// Use mock Git client
			GitClient = NewMockGit()
			mockGit := GitClient.(*MockGit)

			// Configure mock behavior
			mockGit.BranchExistsFunc = func(branchName string) (bool, error) {
				assert.Equal(t, tt.branchName, branchName)
				return tt.branchExists, tt.existsError
			}
			mockGit.CreateBranchFunc = func(branchName string) error {
				assert.Equal(t, tt.branchName, branchName)
				return tt.createError
			}
			mockGit.SwitchBranchFunc = func(branchName string) error {
				assert.Equal(t, tt.branchName, branchName)
				return tt.switchError
			}

			// Test branch creation
			err := CreateBranch(tt.branchName)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
