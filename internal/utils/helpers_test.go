package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		args    []string
		want    string
		wantErr bool
	}{
		{
			name:    "echo command",
			command: "echo",
			args:    []string{"hello"},
			want:    "hello",
			wantErr: false,
		},
		{
			name:    "invalid command",
			command: "nonexistentcommand",
			args:    []string{},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RunCommand(tt.command, tt.args...)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
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
		name    string
		env     map[string]string
		want    string
		wantErr bool
	}{
		{
			name: "default config dir",
			env:  map[string]string{},
			want: filepath.Join(os.Getenv("HOME"), ".tracer"),
		},
		{
			name: "custom config dir",
			env: map[string]string{
				"TRACER_CONFIG_DIR": "", // Will be set in the test
			},
			want: "", // Will be set in the test
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// Reset TestConfigDir
			TestConfigDir = ""
			if tt.name == "custom config dir" {
				// Use a temporary directory for the custom path
				tempDir := t.TempDir()
				customPath := filepath.Join(tempDir, "custom-config")
				TestConfigDir = customPath
				tt.want = customPath
			}

			got, err := GetConfigDir()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
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
