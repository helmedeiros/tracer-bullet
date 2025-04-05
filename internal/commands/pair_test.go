package commands

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var outputBuffer bytes.Buffer

func GetOutput() string {
	return outputBuffer.String()
}

func TestPairCommand(t *testing.T) {
	tmpDir, mockGitClient, originalDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, tmpDir, originalDir)

	// Set up mock git client behavior
	mockGitClient.GetConfigFunc = func(key string) (string, error) {
		switch key {
		case "current.project":
			return "test-project", nil
		case "test-project.user":
			return "john.doe", nil
		case "test-project.pair":
			return "", nil
		}
		return "", nil
	}

	mockGitClient.SetConfigFunc = func(key, value string) error {
		return nil
	}

	// First configure a project and user (required for pair configuration)
	err := configureProject("test-project")
	require.NoError(t, err)
	err = configureUser("john.doe")
	require.NoError(t, err)

	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		errMessage string
		wantOutput string
	}{
		{
			name:       "start with valid partner",
			args:       []string{"start", "jane.doe"},
			wantErr:    false,
			wantOutput: "Started pair programming session with jane.doe\n",
		},
		{
			name:       "start with empty partner",
			args:       []string{"start"},
			wantErr:    true,
			errMessage: "partner name is required",
		},
		{
			name:       "stop pair",
			args:       []string{"stop"},
			wantErr:    false,
			wantOutput: "Stopped pair programming session\n",
		},
		{
			name:       "show status",
			args:       []string{"show"},
			wantErr:    false,
			wantOutput: "\nPair Programming Session:\n  Project: test-project\n  Current User: john.doe\n  Pair Partner: \n\n",
		},
		{
			name:       "invalid command",
			args:       []string{"invalid"},
			wantErr:    true,
			errMessage: "unknown command: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputBuffer.Reset()
			PairCmd.SetOut(&outputBuffer)
			err := PairCmd.RunE(PairCmd, tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Equal(t, tt.errMessage, err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantOutput, GetOutput())
			}
		})
	}
}
