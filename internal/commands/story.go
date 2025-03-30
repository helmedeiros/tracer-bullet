package commands

import (
	"fmt"

	"github.com/helmedeiros/tracer-bullet/internal/config"
	"github.com/helmedeiros/tracer-bullet/internal/story"
	"github.com/spf13/cobra"
)

var StoryCmd = &cobra.Command{
	Use:   "story",
	Short: "Manage stories and their tracking",
	Long:  `Create and manage stories, track their progress, and view associated commits and changes.`,
}

var storyNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new story",
	Long:  `Create a new story with title, description, and other metadata.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get current user from config
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get flag values
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		// Create new story
		s, err := story.NewStory(title, description, cfg.AuthorName)
		if err != nil {
			return err
		}

		// Set tags if provided
		if len(tags) > 0 {
			s.Tags = tags
		}

		// Save story
		if err := s.Save(); err != nil {
			return err
		}

		// Write output to stdout
		fmt.Fprintf(cmd.OutOrStdout(), "Created new story: %s\n", s.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "Title: %s\n", s.Title)
		if s.Description != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Description: %s\n", s.Description)
		}
		if len(s.Tags) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Tags: %v\n", s.Tags)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Author: %s\n", s.Author)
		fmt.Fprintf(cmd.OutOrStdout(), "Status: %s\n", s.Status)

		return nil
	},
}

var storyAfterHashCmd = &cobra.Command{
	Use:   "after-hash",
	Short: "Show stories after a specific commit hash",
	Long:  `Display stories that have been modified after a specific commit hash.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement story history after hash
	},
}

var storyByCmd = &cobra.Command{
	Use:   "by",
	Short: "Show stories by author",
	Long:  `Display stories created or modified by a specific author.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement story filtering by author
	},
}

var storyFilesCmd = &cobra.Command{
	Use:   "files",
	Short: "Show files associated with a story",
	Long:  `Display all files that have been modified as part of a story.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement story files listing
	},
}

var storyCommitsCmd = &cobra.Command{
	Use:   "commits",
	Short: "Show commits associated with a story",
	Long:  `Display all commits that are part of a story's development.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement story commits listing
	},
}

var storyDiaryCmd = &cobra.Command{
	Use:   "diary",
	Short: "Show story development diary",
	Long:  `Display a chronological diary of story development activities.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement story diary
	},
}

var storyDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show story changes",
	Long:  `Display the changes made as part of a story between two points.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement story diff
	},
}

func init() {
	// Add flags to new command
	storyNewCmd.Flags().StringP("title", "t", "", "Story title (required)")
	storyNewCmd.Flags().StringP("description", "d", "", "Story description")
	storyNewCmd.Flags().StringSlice("tags", []string{}, "Story tags")
	storyNewCmd.MarkFlagRequired("title")

	// Add commands to root
	StoryCmd.AddCommand(storyNewCmd)
	StoryCmd.AddCommand(storyAfterHashCmd)
	StoryCmd.AddCommand(storyByCmd)
	StoryCmd.AddCommand(storyFilesCmd)
	StoryCmd.AddCommand(storyCommitsCmd)
	StoryCmd.AddCommand(storyDiaryCmd)
	StoryCmd.AddCommand(storyDiffCmd)
}
