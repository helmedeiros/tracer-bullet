package commands

import (
	"fmt"

	"github.com/helmedeiros/tracer-bullet/internal/config"
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

		// Update config with new values if provided
		if host != "" {
			cfg.JiraHost = host
		}
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
		// TODO: Implement Jira issue creation
		return fmt.Errorf("not implemented yet")
	},
}

var jiraUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing Jira issue",
	Long:  `Update an existing Jira issue with new information.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement Jira issue update
		return fmt.Errorf("not implemented yet")
	},
}

var jiraLinkCmd = &cobra.Command{
	Use:   "link",
	Short: "Link a story to a Jira issue",
	Long:  `Link an existing story to a Jira issue for tracking.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement story-Jira linking
		return fmt.Errorf("not implemented yet")
	},
}

func init() {
	// Configure command flags
	jiraConfigureCmd.Flags().String("host", "", "Jira host URL (required)")
	jiraConfigureCmd.Flags().String("token", "", "Jira API token")
	jiraConfigureCmd.Flags().String("project", "", "Default Jira project key")
	jiraConfigureCmd.Flags().String("user", "", "Jira username/email")

	// Create command flags
	jiraCreateCmd.Flags().StringP("title", "t", "", "Issue title/summary (required)")
	jiraCreateCmd.Flags().StringP("description", "d", "", "Issue description")
	jiraCreateCmd.Flags().String("type", config.DefaultJiraIssueType, "Issue type")
	jiraCreateCmd.Flags().StringSlice("labels", []string{}, "Issue labels")
	jiraCreateCmd.MarkFlagRequired("title")

	// Update command flags
	jiraUpdateCmd.Flags().String("id", "", "Jira issue ID (required)")
	jiraUpdateCmd.Flags().String("status", "", "New issue status")
	jiraUpdateCmd.Flags().String("comment", "", "Add a comment")
	jiraUpdateCmd.MarkFlagRequired("id")

	// Link command flags
	jiraLinkCmd.Flags().String("story", "", "Story ID (required)")
	jiraLinkCmd.Flags().String("issue", "", "Jira issue ID (required)")
	jiraLinkCmd.MarkFlagRequired("story")
	jiraLinkCmd.MarkFlagRequired("issue")

	// Add commands to root
	JiraCmd.AddCommand(jiraConfigureCmd)
	JiraCmd.AddCommand(jiraCreateCmd)
	JiraCmd.AddCommand(jiraUpdateCmd)
	JiraCmd.AddCommand(jiraLinkCmd)
}
