package jira

import (
	"fmt"

	jira "github.com/andygrunwald/go-jira"
)

// MockClient is a mock implementation of the Jira client for testing
type MockClient struct {
	issues      map[string]*jira.Issue
	transitions map[string][]jira.Transition
}

// NewMockClient creates a new mock Jira client
func NewMockClient() *MockClient {
	return &MockClient{
		issues: make(map[string]*jira.Issue),
		transitions: map[string][]jira.Transition{
			"TEST-123": {
				{ID: "1", Name: "In Progress"},
				{ID: "2", Name: "Done"},
			},
		},
	}
}

// CreateIssue creates a mock issue
func (m *MockClient) CreateIssue(title, description, issueType, priority string) (*jira.Issue, error) {
	if title == "" {
		return nil, fmt.Errorf("summary is required")
	}
	if issueType == "" {
		return nil, fmt.Errorf("issuetype is required")
	}

	issue := &jira.Issue{
		Key: "TEST-123",
		Fields: &jira.IssueFields{
			Summary:     title,
			Description: description,
			Type: jira.IssueType{
				Name: issueType,
			},
			Priority: &jira.Priority{
				Name: priority,
			},
		},
	}

	m.issues[issue.Key] = issue
	return issue, nil
}

// GetIssue retrieves a mock issue
func (m *MockClient) GetIssue(issueID string) (*jira.Issue, error) {
	if issueID == "" {
		return nil, fmt.Errorf("issue ID is required")
	}

	issue, exists := m.issues[issueID]
	if !exists {
		return nil, fmt.Errorf("failed to get Jira issue %s: not found", issueID)
	}

	return issue, nil
}

// UpdateIssue updates a mock issue
func (m *MockClient) UpdateIssue(issueID, status, assignee string) error {
	issue, err := m.GetIssue(issueID)
	if err != nil {
		return err
	}

	if status != "" {
		if err := m.updateIssueStatus(issueID, status); err != nil {
			return err
		}
	}

	if assignee != "" {
		issue.Fields.Assignee = &jira.User{Name: assignee}
	}

	return nil
}

// updateIssueStatus updates the status of a mock issue
func (m *MockClient) updateIssueStatus(issueID, status string) error {
	transitions, exists := m.transitions[issueID]
	if !exists {
		return fmt.Errorf("failed to get transitions for issue %s: not found", issueID)
	}

	var transitionID string
	for _, t := range transitions {
		if t.Name == status {
			transitionID = t.ID
			break
		}
	}

	if transitionID == "" {
		return fmt.Errorf("status transition to '%s' not available", status)
	}

	return nil
}
