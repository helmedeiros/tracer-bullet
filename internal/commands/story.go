package commands

import (
	"fmt"
	"time"

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

		// Validate project and user configuration
		if cfg.GitRepo == "" {
			return fmt.Errorf("project not configured. Please run 'tracer configure project' first")
		}
		if cfg.AuthorName == "" {
			return fmt.Errorf("user not configured. Please run 'tracer configure user' first")
		}

		// Get flag values
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		number, _ := cmd.Flags().GetInt("number")

		// Create new story
		s, err := story.NewStoryWithNumber(title, description, cfg.AuthorName, number)
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
		fmt.Fprintf(cmd.OutOrStdout(), "Number: %d\n", s.Number)
		if s.Title != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Title: %s\n", s.Title)
		}
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
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get commit hash from flag
		hash, _ := cmd.Flags().GetString("hash")
		if hash == "" {
			return fmt.Errorf("commit hash is required")
		}

		// Get all stories
		stories, err := story.ListStories()
		if err != nil {
			return fmt.Errorf("failed to list stories: %w", err)
		}

		// Filter stories modified after the given hash
		var modifiedStories []*story.Story
		for _, s := range stories {
			for _, commit := range s.Commits {
				if commit.Hash == hash {
					modifiedStories = append(modifiedStories, s)
					break
				}
			}
		}

		if len(modifiedStories) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No stories found modified after commit %s\n", hash)
			return nil
		}

		// Display stories
		fmt.Fprintf(cmd.OutOrStdout(), "Stories modified after commit %s:\n\n", hash)
		for _, s := range modifiedStories {
			fmt.Fprintf(cmd.OutOrStdout(), "ID: %s\n", s.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "Title: %s\n", s.Title)
			if s.Description != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Description: %s\n", s.Description)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Status: %s\n", s.Status)
			fmt.Fprintf(cmd.OutOrStdout(), "Author: %s\n", s.Author)
			if len(s.Tags) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Tags: %v\n", s.Tags)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "---\n")
		}

		return nil
	},
}

var storyByCmd = &cobra.Command{
	Use:   "by",
	Short: "Show stories by author",
	Long:  `Display stories created or modified by a specific author.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get author from flag
		author, _ := cmd.Flags().GetString("author")
		if author == "" {
			return fmt.Errorf("author is required")
		}

		// Get all stories
		stories, err := story.ListStories()
		if err != nil {
			return fmt.Errorf("failed to list stories: %w", err)
		}

		// Filter stories by author
		var authorStories []*story.Story
		for _, s := range stories {
			if s.Author == author {
				authorStories = append(authorStories, s)
			}
		}

		if len(authorStories) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No stories found for author %s\n", author)
			return nil
		}

		// Display stories
		fmt.Fprintf(cmd.OutOrStdout(), "Stories by %s:\n\n", author)
		for _, s := range authorStories {
			fmt.Fprintf(cmd.OutOrStdout(), "ID: %s\n", s.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "Title: %s\n", s.Title)
			if s.Description != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Description: %s\n", s.Description)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Status: %s\n", s.Status)
			if len(s.Tags) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Tags: %v\n", s.Tags)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", s.CreatedAt.Format(time.RFC3339))
			fmt.Fprintf(cmd.OutOrStdout(), "---\n")
		}

		return nil
	},
}

var storyFilesCmd = &cobra.Command{
	Use:   "files",
	Short: "Show files associated with a story",
	Long:  `Display all files that have been modified as part of a story.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get story ID from flag
		storyID, _ := cmd.Flags().GetString("id")
		if storyID == "" {
			return fmt.Errorf("story ID is required")
		}

		// Load the story
		s, err := story.LoadStory(storyID)
		if err != nil {
			return fmt.Errorf("failed to load story: %w", err)
		}

		// Get files
		files := s.GetFiles()
		if len(files) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No files found for story %s\n", storyID)
			return nil
		}

		// Display files
		fmt.Fprintf(cmd.OutOrStdout(), "Files for story %s (%s):\n\n", s.ID, s.Title)
		for _, file := range files {
			fmt.Fprintf(cmd.OutOrStdout(), "Path: %s\n", file.Path)
			fmt.Fprintf(cmd.OutOrStdout(), "Status: %s\n", file.Status)
			fmt.Fprintf(cmd.OutOrStdout(), "Modified: %s\n", file.Timestamp.Format(time.RFC3339))
			fmt.Fprintf(cmd.OutOrStdout(), "---\n")
		}

		return nil
	},
}

var storyCommitsCmd = &cobra.Command{
	Use:   "commits",
	Short: "Show commits associated with a story",
	Long:  `Display all commits that are part of a story's development.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get story ID from flag
		storyID, _ := cmd.Flags().GetString("id")
		if storyID == "" {
			return fmt.Errorf("story ID is required")
		}

		// Load the story
		s, err := story.LoadStory(storyID)
		if err != nil {
			return fmt.Errorf("failed to load story: %w", err)
		}

		// Get commits
		commits := s.GetCommits()
		if len(commits) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No commits found for story %s\n", storyID)
			return nil
		}

		// Display commits
		fmt.Fprintf(cmd.OutOrStdout(), "Commits for story %s (%s):\n\n", s.ID, s.Title)
		for _, commit := range commits {
			fmt.Fprintf(cmd.OutOrStdout(), "Hash: %s\n", commit.Hash)
			fmt.Fprintf(cmd.OutOrStdout(), "Author: %s\n", commit.Author)
			fmt.Fprintf(cmd.OutOrStdout(), "Date: %s\n", commit.Timestamp.Format(time.RFC3339))
			fmt.Fprintf(cmd.OutOrStdout(), "Message: %s\n", commit.Message)
			fmt.Fprintf(cmd.OutOrStdout(), "---\n")
		}

		return nil
	},
}

var storyDiaryCmd = &cobra.Command{
	Use:   "diary",
	Short: "Show story development diary",
	Long:  `Display a chronological diary of story development activities.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get story ID from flag
		storyID, _ := cmd.Flags().GetString("id")
		if storyID == "" {
			return fmt.Errorf("story ID is required")
		}

		// Load the story
		s, err := story.LoadStory(storyID)
		if err != nil {
			return fmt.Errorf("failed to load story: %w", err)
		}

		// Get time range from flags
		since, _ := cmd.Flags().GetString("since")
		until, _ := cmd.Flags().GetString("until")

		// Parse time range
		var startTime, endTime time.Time
		if since != "" {
			startTime, err = time.Parse(time.RFC3339, since)
			if err != nil {
				return fmt.Errorf("invalid since time format: %w", err)
			}
		} else {
			startTime = s.CreatedAt
		}

		if until != "" {
			endTime, err = time.Parse(time.RFC3339, until)
			if err != nil {
				return fmt.Errorf("invalid until time format: %w", err)
			}
		} else {
			endTime = time.Now()
		}

		// Display story info
		fmt.Fprintf(cmd.OutOrStdout(), "Story Diary: %s (%s)\n\n", s.Title, s.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "Time Range: %s to %s\n\n",
			startTime.Format(time.RFC3339),
			endTime.Format(time.RFC3339))

		// Display commits in chronological order
		if len(s.Commits) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Commits:\n")
			for _, commit := range s.Commits {
				if commit.Timestamp.After(startTime) && commit.Timestamp.Before(endTime) {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", commit.Timestamp.Format(time.RFC3339))
					fmt.Fprintf(cmd.OutOrStdout(), "  Hash: %s\n", commit.Hash)
					fmt.Fprintf(cmd.OutOrStdout(), "  Author: %s\n", commit.Author)
					fmt.Fprintf(cmd.OutOrStdout(), "  Message: %s\n", commit.Message)
					fmt.Fprintf(cmd.OutOrStdout(), "  ---\n")
				}
			}
		}

		// Display file changes in chronological order
		if len(s.Files) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "\nFile Changes:\n")
			for _, file := range s.Files {
				if file.Timestamp.After(startTime) && file.Timestamp.Before(endTime) {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", file.Timestamp.Format(time.RFC3339))
					fmt.Fprintf(cmd.OutOrStdout(), "  Path: %s\n", file.Path)
					fmt.Fprintf(cmd.OutOrStdout(), "  Status: %s\n", file.Status)
					fmt.Fprintf(cmd.OutOrStdout(), "  ---\n")
				}
			}
		}

		return nil
	},
}

var storyDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show story changes",
	Long:  `Display the changes made as part of a story between two points.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get story ID from flag
		storyID, _ := cmd.Flags().GetString("id")
		if storyID == "" {
			return fmt.Errorf("story ID is required")
		}

		// Get from and to points from flags
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")

		// Load the story
		s, err := story.LoadStory(storyID)
		if err != nil {
			return fmt.Errorf("failed to load story: %w", err)
		}

		// If no from/to points specified, use story creation and last update
		if from == "" {
			from = s.CreatedAt.Format(time.RFC3339)
		}
		if to == "" {
			to = s.UpdatedAt.Format(time.RFC3339)
		}

		// Display story info
		fmt.Fprintf(cmd.OutOrStdout(), "Story Changes: %s (%s)\n\n", s.Title, s.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "Time Range: %s to %s\n\n", from, to)

		// Display commits in the range
		if len(s.Commits) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Commits:\n")
			for _, commit := range s.Commits {
				commitTime := commit.Timestamp.Format(time.RFC3339)
				if commitTime >= from && commitTime <= to {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", commitTime)
					fmt.Fprintf(cmd.OutOrStdout(), "  Hash: %s\n", commit.Hash)
					fmt.Fprintf(cmd.OutOrStdout(), "  Author: %s\n", commit.Author)
					fmt.Fprintf(cmd.OutOrStdout(), "  Message: %s\n", commit.Message)
					fmt.Fprintf(cmd.OutOrStdout(), "  ---\n")
				}
			}
		}

		// Display file changes in the range
		if len(s.Files) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "\nFile Changes:\n")
			for _, file := range s.Files {
				fileTime := file.Timestamp.Format(time.RFC3339)
				if fileTime >= from && fileTime <= to {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", fileTime)
					fmt.Fprintf(cmd.OutOrStdout(), "  Path: %s\n", file.Path)
					fmt.Fprintf(cmd.OutOrStdout(), "  Status: %s\n", file.Status)
					fmt.Fprintf(cmd.OutOrStdout(), "  ---\n")
				}
			}
		}

		return nil
	},
}

func init() {
	// Add commands to root
	StoryCmd.AddCommand(storyNewCmd)
	StoryCmd.AddCommand(storyAfterHashCmd)
	StoryCmd.AddCommand(storyByCmd)
	StoryCmd.AddCommand(storyFilesCmd)
	StoryCmd.AddCommand(storyCommitsCmd)
	StoryCmd.AddCommand(storyDiaryCmd)
	StoryCmd.AddCommand(storyDiffCmd)

	// Add new command flags
	storyNewCmd.Flags().StringP("title", "t", "", "Story title")
	storyNewCmd.Flags().StringP("description", "d", "", "Story description")
	storyNewCmd.Flags().StringSlice("tags", []string{}, "Story tags")
	storyNewCmd.Flags().IntP("number", "n", 0, "Story number")
	if err := storyNewCmd.MarkFlagRequired("number"); err != nil {
		panic(fmt.Sprintf("failed to mark number flag as required: %v", err))
	}

	// Add after-hash command flags
	storyAfterHashCmd.Flags().String("hash", "", "Commit hash")

	// Add by command flags
	storyByCmd.Flags().StringP("author", "a", "", "Story author")

	// Add files command flags
	storyFilesCmd.Flags().StringP("id", "i", "", "Story ID")

	// Add commits command flags
	storyCommitsCmd.Flags().StringP("id", "i", "", "Story ID")

	// Add diary command flags
	storyDiaryCmd.Flags().StringP("id", "i", "", "Story ID")
	storyDiaryCmd.Flags().String("since", "", "Start time (RFC3339 format)")
	storyDiaryCmd.Flags().String("until", "", "End time (RFC3339 format)")

	// Add diff command flags
	storyDiffCmd.Flags().StringP("id", "i", "", "Story ID")
	storyDiffCmd.Flags().String("from", "", "Start point (RFC3339 format)")
	storyDiffCmd.Flags().String("to", "", "End point (RFC3339 format)")

	// Handle required flags
	requiredFlags := map[*cobra.Command][]string{
		storyNewCmd:       {"number"},
		storyAfterHashCmd: {"hash"},
		storyByCmd:        {"author"},
		storyFilesCmd:     {"id"},
		storyCommitsCmd:   {"id"},
		storyDiaryCmd:     {"id"},
		storyDiffCmd:      {"id"},
	}

	for cmd, flags := range requiredFlags {
		for _, flag := range flags {
			if err := cmd.MarkFlagRequired(flag); err != nil {
				panic(fmt.Sprintf("failed to mark %s flag as required for %s: %v", flag, cmd.Name(), err))
			}
		}
	}
}
