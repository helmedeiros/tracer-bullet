package utils

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	// DefaultDirPerm is the default permission for directories
	DefaultDirPerm = 0755
	// DefaultFilePerm is the default permission for files
	DefaultFilePerm = 0600
)

var (
	// TestConfigDir is used to override the config directory during tests
	TestConfigDir string

	// GitClient is the global git client, can be replaced with a mock for testing
	GitClient GitOperations = NewRealGit()
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
	if TestConfigDir != "" {
		return TestConfigDir, nil
	}

	homeDir, err := GetHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".tracer")
	if err := EnsureDir(configDir); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return configDir, nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// RunCommand executes a command and returns its output
func RunCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", exitErr
		}
		return "", err
	}
	return string(output), nil
}

// GenerateID generates a unique ID using random bytes
func GenerateID() string {
	b := make([]byte, 16)
	n, err := rand.Read(b)
	if err != nil || n != 16 {
		// If we can't generate random bytes, use timestamp as fallback
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", b)
}

// RunGitInit initializes a new git repository in the current directory
func RunGitInit() error {
	return GitClient.Init()
}

// RunGitConfig sets a git configuration value
func RunGitConfig(key, value string) error {
	return GitClient.SetConfig(key, value)
}

// GetGitRoot returns the root directory of the git repository
func GetGitRoot() (string, error) {
	return GitClient.GetGitRoot()
}

// GetRepoConfigDir returns the repository-specific configuration directory
func GetRepoConfigDir() (string, error) {
	gitRoot, err := GetGitRoot()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(gitRoot, ".tracer")

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(configDir, DefaultDirPerm); err != nil {
		return "", fmt.Errorf("failed to create repo config directory: %w", err)
	}

	// Ensure the directory is writable
	if err := checkDirWritable(configDir); err != nil {
		return "", fmt.Errorf("repo config directory is not writable: %w", err)
	}

	return configDir, nil
}

// checkDirWritable checks if a directory is writable by attempting to create a temporary file
func checkDirWritable(dir string) error {
	// First check if the path exists and is a directory
	if info, err := os.Stat(dir); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("path exists but is not a directory: %s", dir)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check directory: %w", err)
	}

	// If the directory doesn't exist, try to create it
	if err := os.MkdirAll(dir, DefaultDirPerm); err != nil {
		return fmt.Errorf("directory does not exist and cannot be created: %w", err)
	}

	// Try to create a temporary file to check if the directory is writable
	tmpFile := filepath.Join(dir, ".tracer_test_"+GenerateID())
	if err := os.WriteFile(tmpFile, []byte("test"), DefaultFilePerm); err != nil {
		return fmt.Errorf("directory is not writable: %w", err)
	}

	// Clean up the temporary file
	if err := os.Remove(tmpFile); err != nil {
		return fmt.Errorf("failed to remove temporary file: %w", err)
	}

	return nil
}

// CreateBranch creates a new git branch and switches to it
func CreateBranch(branchName string) error {
	// Check if branch already exists
	exists, err := GitClient.BranchExists(branchName)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}

	if exists {
		// If branch exists, just switch to it
		return GitClient.SwitchBranch(branchName)
	}

	// Create and switch to new branch
	return GitClient.CreateBranch(branchName)
}

// ToKebabCase converts a string to kebab-case format
func ToKebabCase(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace special characters with dashes
	replacements := []string{
		" ", "-",
		"_", "-",
		".", "-",
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "-",
		"<", "-",
		">", "-",
		"|", "-",
		"#", "-",
	}

	replacer := strings.NewReplacer(replacements...)
	s = replacer.Replace(s)

	// Remove any consecutive dashes
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}

	// Remove leading and trailing dashes
	return strings.Trim(s, "-")
}

// GenerateBranchName generates a kebab-case branch name from a story title or ID
func GenerateBranchName(title string, id string, number int, project string) string {
	// Use title if available, otherwise use ID
	name := title
	if name == "" {
		name = id
	}

	// Convert to kebab case
	name = ToKebabCase(name)

	// If number is provided, prepend it to the name
	if number > 0 {
		name = fmt.Sprintf("%d-%s", number, name)
	}

	// If project is provided, prepend it to the name
	if project != "" {
		name = fmt.Sprintf("%s-%s", project, name)
	}

	// Prepend features/ to the branch name
	return "features/" + name
}

// GetProjectName returns the current project name from git config
func GetProjectName() (string, error) {
	return GitClient.GetConfig("current.project")
}
