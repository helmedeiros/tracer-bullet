package commands

import (
	"fmt"

	"github.com/helmedeiros/tracer-bullet/internal/config"
	"github.com/helmedeiros/tracer-bullet/internal/jira"
	"github.com/helmedeiros/tracer-bullet/internal/story"
	"github.com/spf13/cobra"
)

var JiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Manage Jira integration",
	Long:  `Configure and interact with Jira, including creating and updating issues.`,
}

var jiraConfigureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Jira settings",
	Long:  `Configure Jira settings including host, project, and authentication.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load existing config
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get flag values
		host, _ := cmd.Flags().GetString("host")
		token, _ := cmd.Flags().GetString("token")
		project, _ := cmd.Flags().GetString("project")
		user, _ := cmd.Flags().GetString("user")

		// Validate host parameter
		if host == "" {
			return fmt.Errorf("host cannot be empty")
		}

		// Update config with new values if provided
		cfg.JiraHost = host
		if token != "" {
			cfg.JiraToken = token
		}
		if project != "" {
			cfg.JiraProject = project
		}
		if user != "" {
			cfg.JiraUser = user
		}

		// Save updated config
		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		// Print current configuration
		fmt.Fprintf(cmd.OutOrStdout(), "Jira configuration updated:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "Host: %s\n", cfg.JiraHost)
		fmt.Fprintf(cmd.OutOrStdout(), "Project: %s\n", cfg.JiraProject)
		fmt.Fprintf(cmd.OutOrStdout(), "User: %s\n", cfg.JiraUser)
		if cfg.JiraToken != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Token: [CONFIGURED]\n")
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Token: [NOT CONFIGURED]\n")
		}

		return nil
	},
}

var jiraCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Jira issue",
	Long:  `Create a new Jira issue with title, description, and other metadata.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Create Jira client
		client, err := jira.NewClient(cfg)
		if err != nil {
			return fmt.Errorf("failed to create Jira client: %w", err)
		}

		// Get flag values
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		issueType, _ := cmd.Flags().GetString("type")
		priority, _ := cmd.Flags().GetString("priority")

		// Create the issue
		issue, err := client.CreateIssue(title, description, issueType, priority)
		if err != nil {
			return fmt.Errorf("failed to create issue: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Created Jira issue: %s\n", issue.Key)
		fmt.Fprintf(cmd.OutOrStdout(), "URL: %s/browse/%s\n", cfg.JiraHost, issue.Key)
		return nil
	},
}

var jiraUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing Jira issue",
	Long:  `Update an existing Jira issue with new information.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Create Jira client
		client, err := jira.NewClient(cfg)
		if err != nil {
			return fmt.Errorf("failed to create Jira client: %w", err)
		}

		// Get flag values
		issueID, _ := cmd.Flags().GetString("id")
		status, _ := cmd.Flags().GetString("status")
		assignee, _ := cmd.Flags().GetString("assignee")

		// Update the issue
		err = client.UpdateIssue(issueID, status, assignee)
		if err != nil {
			return fmt.Errorf("failed to update issue: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Updated Jira issue: %s\n", issueID)
		fmt.Fprintf(cmd.OutOrStdout(), "URL: %s/browse/%s\n", cfg.JiraHost, issueID)
		return nil
	},
}

var jiraLinkCmd = &cobra.Command{
	Use:   "link",
	Short: "Link a story to a Jira issue",
	Long:  `Link an existing story to a Jira issue for tracking.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Create Jira client
		client, err := jira.NewClient(cfg)
		if err != nil {
			return fmt.Errorf("failed to create Jira client: %w", err)
		}

		// Get flag values
		storyID, _ := cmd.Flags().GetString("story")
		issueID, _ := cmd.Flags().GetString("issue")

		// Verify the Jira issue exists
		issue, err := client.GetIssue(issueID)
		if err != nil {
			return fmt.Errorf("failed to get Jira issue: %w", err)
		}

		// Load the story
		s, err := story.LoadStory(storyID)
		if err != nil {
			return fmt.Errorf("failed to load story: %w", err)
		}

		// Update story with Jira issue key
		s.JiraKey = issue.Key
		if err := story.SaveStory(s); err != nil {
			return fmt.Errorf("failed to save story: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Linked story %s to Jira issue %s\n", storyID, issue.Key)
		fmt.Fprintf(cmd.OutOrStdout(), "URL: %s/browse/%s\n", cfg.JiraHost, issue.Key)
		return nil
	},
}

func init() {
	// Add commands to root
	JiraCmd.AddCommand(jiraConfigureCmd)
	JiraCmd.AddCommand(jiraCreateCmd)
	JiraCmd.AddCommand(jiraUpdateCmd)
	JiraCmd.AddCommand(jiraLinkCmd)

	// Add configure command flags
	jiraConfigureCmd.Flags().String("host", "", "Jira host URL")
	jiraConfigureCmd.Flags().String("token", "", "Jira API token")
	jiraConfigureCmd.Flags().String("project", "", "Default Jira project key")
	jiraConfigureCmd.Flags().String("user", "", "Jira username/email")

	// Add create command flags
	jiraCreateCmd.Flags().String("title", "", "Issue title")
	jiraCreateCmd.Flags().String("description", "", "Issue description")
	jiraCreateCmd.Flags().String("type", "Task", "Issue type (default: Task)")
	jiraCreateCmd.Flags().String("priority", "Medium", "Issue priority (default: Medium)")

	// Add update command flags
	jiraUpdateCmd.Flags().String("id", "", "Issue ID")
	jiraUpdateCmd.Flags().String("status", "", "New status")
	jiraUpdateCmd.Flags().String("assignee", "", "New assignee")

	// Add link command flags
	jiraLinkCmd.Flags().String("story", "", "Story ID")
	jiraLinkCmd.Flags().String("issue", "", "Jira issue ID")

	// Handle required flags
	requiredFlags := map[*cobra.Command][]string{
		jiraCreateCmd: {"title"},
		jiraUpdateCmd: {"id"},
		jiraLinkCmd:   {"story", "issue"},
	}

	for cmd, flags := range requiredFlags {
		for _, flag := range flags {
			if err := cmd.MarkFlagRequired(flag); err != nil {
				panic(fmt.Sprintf("failed to mark %s flag as required for %s: %v", flag, cmd.Name(), err))
			}
		}
	}
}
