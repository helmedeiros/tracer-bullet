package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/helmedeiros/tracer-bullet/internal/utils"
	"github.com/stretchr/testify/require"
)

// setupTestEnvironment creates a temporary test environment and returns:
// - tmpDir: the temporary directory
// - repoDir: the test repository directory
// - originalDir: the original working directory
func setupTestEnvironment(t *testing.T) (string, string, string) {
	// Create a temporary directory for testing with a unique name
	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("tracer-test-%s-*", t.Name()))
	require.NoError(t, err)

	// Get the original directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	// Create test repository directory
	repoDir := filepath.Join(tmpDir, "test-repo")
	err = os.MkdirAll(repoDir, 0755)
	require.NoError(t, err)

	// Change to the test repository directory
	err = os.Chdir(repoDir)
	require.NoError(t, err)

	// Initialize git repository
	err = utils.RunGitInit()
	require.NoError(t, err)

	// Create config directory in temporary directory
	configDir := filepath.Join(tmpDir, ".tracer")
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	// Override config directory for tests
	utils.TestConfigDir = configDir

	return tmpDir, repoDir, originalDir
}

// cleanupTestEnvironment cleans up the test environment
func cleanupTestEnvironment(t *testing.T, tmpDir, originalDir string) {
	// Reset config directory override
	utils.TestConfigDir = ""

	// Change back to original directory
	if _, err := os.Stat(originalDir); err == nil {
		_ = os.Chdir(originalDir)
	}

	// Remove temporary directory
	if tmpDir != "" {
		_ = os.RemoveAll(tmpDir)
	}
}
