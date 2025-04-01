package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// DefaultDirPerm is the default permission for directories
	DefaultDirPerm = 0755
	// DefaultFilePerm is the default permission for files
	DefaultFilePerm = 0600
)

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, DefaultDirPerm)
	}
	return nil
}

// GetHomeDir returns the user's home directory
func GetHomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return home, nil
}

// GetConfigDir returns the tracer configuration directory
func GetConfigDir() (string, error) {
	home, err := GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".tracer"), nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// RunCommand executes a shell command and returns its output
func RunCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run command %s: %w", command, err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GenerateID generates a unique ID using random bytes
func GenerateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:8]
}

// RunGitInit initializes a new git repository in the current directory
func RunGitInit() error {
	_, err := RunCommand("git", "init")
	if err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Configure git user for tests
	if err := RunGitConfig("user.name", "Test User"); err != nil {
		return err
	}
	if err := RunGitConfig("user.email", "test@example.com"); err != nil {
		return err
	}

	return nil
}

// RunGitConfig sets a git configuration value
func RunGitConfig(key, value string) error {
	_, err := RunCommand("git", "config", "--local", key, value)
	if err != nil {
		return fmt.Errorf("failed to set git config %s: %w", key, err)
	}
	return nil
}
