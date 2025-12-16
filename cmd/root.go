package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hiroyannnn/gh-pr-inbox/internal/github"
	"github.com/hiroyannnn/gh-pr-inbox/internal/inbox"
	"github.com/spf13/cobra"
)

var (
	repository string
	prNumber   int
	format     string
)

var rootCmd = &cobra.Command{
	Use:   "pr-inbox [PR_NUMBER]",
	Short: "Collect and organize PR review comments into an actionable inbox",
	Long: `gh pr-inbox collects PR review comments, groups them into threads,
filters unresolved issues, prioritizes them, and outputs a clean,
actionable "Inbox" for the PR author.

The purpose is to reduce review noise and make it easy to decide what to fix next.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInbox,
}

func init() {
	rootCmd.Flags().StringVarP(&repository, "repo", "R", "", "Repository in OWNER/REPO format")
	rootCmd.Flags().IntVarP(&prNumber, "pr", "p", 0, "Pull request number")
	rootCmd.Flags().StringVarP(&format, "format", "f", "default", "Output format (default, json)")
}

func Execute() error {
	return rootCmd.Execute()
}

func runInbox(cmd *cobra.Command, args []string) error {
	// Determine PR number
	if len(args) > 0 {
		fmt.Sscanf(args[0], "%d", &prNumber)
	}

	if prNumber == 0 {
		// Try to get PR number from current branch
		var err error
		prNumber, err = getCurrentPRNumber()
		if err != nil || prNumber == 0 {
			return fmt.Errorf("PR number required: specify with argument, --pr flag, or run from a PR branch")
		}
	}

	// Determine repository
	if repository == "" {
		var err error
		repository, err = getCurrentRepository()
		if err != nil {
			return fmt.Errorf("repository required: specify with --repo flag or run from a git repository")
		}
	}

	// Fetch PR data
	client := github.NewClient(repository)
	pr, err := client.GetPullRequest(prNumber)
	if err != nil {
		return fmt.Errorf("failed to fetch PR: %w", err)
	}

	// Collect review comments
	comments, err := client.GetReviewComments(prNumber)
	if err != nil {
		return fmt.Errorf("failed to fetch review comments: %w", err)
	}

	// Process and organize into inbox
	processor := inbox.NewProcessor()
	inboxItems := processor.Process(pr, comments)

	// Output results
	if format == "json" {
		return outputJSON(inboxItems)
	}
	return outputDefault(inboxItems)
}

func getCurrentRepository() (string, error) {
	cmd := exec.Command("gh", "repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getCurrentPRNumber() (int, error) {
	cmd := exec.Command("gh", "pr", "view", "--json", "number", "-q", ".number")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	var prNum int
	fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &prNum)
	return prNum, nil
}

func outputJSON(items []inbox.InboxItem) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(items)
}

func outputDefault(items []inbox.InboxItem) error {
	if len(items) == 0 {
		fmt.Println("🎉 Your PR inbox is empty! No unresolved review comments.")
		return nil
	}

	fmt.Printf("📬 PR Review Inbox (%d items)\n", len(items))
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	for i, item := range items {
		fmt.Printf("[%d] %s\n", i+1, item.Priority)
		fmt.Printf("    Thread: %s\n", item.ThreadID)
		fmt.Printf("    File: %s:%d\n", item.FilePath, item.LineNumber)
		fmt.Printf("    Author: %s\n", item.Author)
		fmt.Printf("    Comment: %s\n", truncate(item.Body, 80))
		if item.UnresolvedCount > 1 {
			fmt.Printf("    Thread has %d unresolved comments\n", item.UnresolvedCount)
		}
		fmt.Println()
	}

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
