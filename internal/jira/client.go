package jira

import (
	"fmt"

	jira "github.com/andygrunwald/go-jira"
	"github.com/helmedeiros/tracer-bullet/internal/config"
)

// Client represents a Jira client wrapper
type Client struct {
	client *jira.Client
	cfg    *config.Config
}

// NewClient creates a new Jira client using the provided configuration
func NewClient(cfg *config.Config) (*Client, error) {
	if cfg.JiraHost == "" {
		return nil, fmt.Errorf("Jira host is not configured")
	}
	if cfg.JiraToken == "" {
		return nil, fmt.Errorf("Jira token is not configured")
	}

	tp := jira.BasicAuthTransport{
		Username: cfg.JiraUser,
		Password: cfg.JiraToken,
	}

	client, err := jira.NewClient(tp.Client(), cfg.JiraHost)
	if err != nil {
		return nil, fmt.Errorf("failed to create Jira client: %w", err)
	}

	return &Client{
		client: client,
		cfg:    cfg,
	}, nil
}

// CreateIssue creates a new Jira issue
func (c *Client) CreateIssue(title, description, issueType, priority string) (*jira.Issue, error) {
	i := jira.Issue{
		Fields: &jira.IssueFields{
			Project: jira.Project{
				Key: c.cfg.JiraProject,
			},
			Type: jira.IssueType{
				Name: issueType,
			},
			Summary:     title,
			Description: description,
		},
	}

	if priority != "" {
		i.Fields.Priority = &jira.Priority{Name: priority}
	}

	issue, _, err := c.client.Issue.Create(&i)
	if err != nil {
		return nil, fmt.Errorf("failed to create Jira issue: %w", err)
	}

	return issue, nil
}

func (c *Client) updateIssueStatus(issueID, status string) error {
	// Get available transitions
	transitions, _, err := c.client.Issue.GetTransitions(issueID)
	if err != nil {
		return fmt.Errorf("failed to get transitions for issue %s: %w", issueID, err)
	}

	// Find the transition for the requested status
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

	_, err = c.client.Issue.DoTransition(issueID, transitionID)
	if err != nil {
		return fmt.Errorf("failed to transition issue %s to %s: %w", issueID, status, err)
	}

	return nil
}

func (c *Client) UpdateIssue(issueID, status, assignee string) error {
	issue, _, err := c.client.Issue.Get(issueID, nil)
	if err != nil {
		return fmt.Errorf("failed to get Jira issue %s: %w", issueID, err)
	}

	if status != "" {
		if err := c.updateIssueStatus(issueID, status); err != nil {
			return err
		}
	}

	if assignee != "" {
		issue.Fields.Assignee = &jira.User{Name: assignee}
		_, _, err = c.client.Issue.Update(issue)
		if err != nil {
			return fmt.Errorf("failed to update assignee for issue %s: %w", issueID, err)
		}
	}

	return nil
}

// GetIssue retrieves a Jira issue by ID
func (c *Client) GetIssue(issueID string) (*jira.Issue, error) {
	issue, _, err := c.client.Issue.Get(issueID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Jira issue %s: %w", issueID, err)
	}
	return issue, nil
}
