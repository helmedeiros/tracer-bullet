package jira

import (
	"testing"

	"github.com/helmedeiros/tracer-bullet/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &config.Config{
				JiraHost:    "https://jira.example.com",
				JiraToken:   "token123",
				JiraUser:    "user@example.com",
				JiraProject: "TEST",
			},
			expectError: false,
		},
		{
			name: "missing host",
			config: &config.Config{
				JiraToken:   "token123",
				JiraUser:    "user@example.com",
				JiraProject: "TEST",
			},
			expectError: true,
		},
		{
			name: "missing token",
			config: &config.Config{
				JiraHost:    "https://jira.example.com",
				JiraUser:    "user@example.com",
				JiraProject: "TEST",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.config, client.cfg)
			}
		})
	}
}
