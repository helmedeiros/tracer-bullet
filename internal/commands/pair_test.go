package commands

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPairCommand(t *testing.T) {
	tmpDir, originalDir := setupTestEnvironment(t)
	defer func() {
		err := os.Chdir(originalDir)
		require.NoError(t, err)
		os.RemoveAll(tmpDir)
	}()

	// First configure a project and user (required for pair command)
	err := configureProject("test-project")
	require.NoError(t, err)
	err = configureUser("john.doe")
	require.NoError(t, err)

	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectedOutput string
	}{
		{
			name:           "start pair session with valid partner",
			args:           []string{"start", "jane.doe"},
			expectError:    false,
			expectedOutput: "Started pair programming session with jane.doe\n",
		},
		{
			name:           "start pair session with empty partner",
			args:           []string{"start", ""},
			expectError:    true,
			expectedOutput: "",
		},
		{
			name:           "start pair session without partner",
			args:           []string{"start"},
			expectError:    true,
			expectedOutput: "",
		},
		{
			name:           "stop pair session",
			args:           []string{"stop"},
			expectError:    false,
			expectedOutput: "Stopped pair programming session\n",
		},
		{
			name:           "show pair status",
			args:           []string{"status"},
			expectError:    false,
			expectedOutput: "No active pair programming session\n",
		},
		{
			name:           "invalid command",
			args:           []string{"invalid"},
			expectError:    true,
			expectedOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PairCmd.RunE(PairCmd, tt.args)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, GetOutput())
			}
		})
	}
}
