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

func TestCreateIssue(t *testing.T) {
	tests := []struct {
		name          string
		title         string
		description   string
		issueType     string
		priority      string
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid issue",
			title:       "Test Issue",
			description: "Test Description",
			issueType:   "Task",
			priority:    "High",
			expectError: false,
		},
		{
			name:          "empty title",
			title:         "",
			description:   "Test Description",
			issueType:     "Task",
			priority:      "High",
			expectError:   true,
			errorContains: "summary is required",
		},
		{
			name:          "empty issue type",
			title:         "Test Issue",
			description:   "Test Description",
			issueType:     "",
			priority:      "High",
			expectError:   true,
			errorContains: "issuetype is required",
		},
	}

	mockClient := NewMockClient()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue, err := mockClient.CreateIssue(tt.title, tt.description, tt.issueType, tt.priority)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, issue)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, issue)
				assert.Equal(t, tt.title, issue.Fields.Summary)
				assert.Equal(t, tt.description, issue.Fields.Description)
				assert.Equal(t, tt.issueType, issue.Fields.Type.Name)
				assert.Equal(t, tt.priority, issue.Fields.Priority.Name)
			}
		})
	}
}

func TestUpdateIssue(t *testing.T) {
	tests := []struct {
		name          string
		issueID       string
		status        string
		assignee      string
		expectError   bool
		errorContains string
	}{
		{
			name:        "update status only",
			issueID:     "TEST-123",
			status:      "In Progress",
			expectError: false,
		},
		{
			name:        "update assignee only",
			issueID:     "TEST-123",
			assignee:    "user@example.com",
			expectError: false,
		},
		{
			name:        "update both status and assignee",
			issueID:     "TEST-123",
			status:      "Done",
			assignee:    "user@example.com",
			expectError: false,
		},
		{
			name:          "invalid issue ID",
			issueID:       "INVALID",
			expectError:   true,
			errorContains: "failed to get Jira issue",
		},
	}

	mockClient := NewMockClient()

	// Create a test issue first
	_, err := mockClient.CreateIssue("Test Issue", "Test Description", "Task", "High")
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mockClient.UpdateIssue(tt.issueID, tt.status, tt.assignee)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetIssue(t *testing.T) {
	tests := []struct {
		name          string
		issueID       string
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid issue",
			issueID:     "TEST-123",
			expectError: false,
		},
		{
			name:          "invalid issue ID",
			issueID:       "INVALID",
			expectError:   true,
			errorContains: "failed to get Jira issue",
		},
		{
			name:          "empty issue ID",
			issueID:       "",
			expectError:   true,
			errorContains: "issue ID is required",
		},
	}

	mockClient := NewMockClient()

	// Create a test issue first
	_, err := mockClient.CreateIssue("Test Issue", "Test Description", "Task", "High")
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue, err := mockClient.GetIssue(tt.issueID)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, issue)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, issue)
				assert.Equal(t, tt.issueID, issue.Key)
			}
		})
	}
}

func TestUpdateIssueStatus(t *testing.T) {
	tests := []struct {
		name          string
		issueID       string
		status        string
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid status transition",
			issueID:     "TEST-123",
			status:      "In Progress",
			expectError: false,
		},
		{
			name:          "invalid status",
			issueID:       "TEST-123",
			status:        "InvalidStatus",
			expectError:   true,
			errorContains: "status transition to 'InvalidStatus' not available",
		},
		{
			name:          "invalid issue ID",
			issueID:       "INVALID",
			status:        "In Progress",
			expectError:   true,
			errorContains: "failed to get transitions",
		},
	}

	mockClient := NewMockClient()

	// Create a test issue first
	_, err := mockClient.CreateIssue("Test Issue", "Test Description", "Task", "High")
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mockClient.updateIssueStatus(tt.issueID, tt.status)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
