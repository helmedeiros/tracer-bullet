package utils

import (
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
