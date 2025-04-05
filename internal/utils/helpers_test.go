package utils

import (
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
	}{
		{
			name:       "echo command",
			command:    "echo",
			args:       []string{"hello world"},
			wantErr:    false,
			wantOutput: "hello world\n",
		},
		{
			name:       "invalid command",
			command:    "nonexistentcommand",
			args:       []string{},
			wantErr:    true,
			wantOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := RunCommand(tt.command, tt.args...)
			if tt.wantErr {
				assert.Error(t, err)
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

func TestEnsureDir(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "create new directory",
			path:    filepath.Join(t.TempDir(), "newdir"),
			wantErr: false,
		},
		{
			name:    "existing directory",
			path:    t.TempDir(),
			wantErr: false,
		},
		{
			name:    "invalid path",
			path:    "/invalid/path/with/permission/denied",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EnsureDir(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.DirExists(t, tt.path)
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
