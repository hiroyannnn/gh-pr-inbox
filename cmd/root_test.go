package cmd

import (
	"testing"

	"github.com/hiroyannnn/gh-pr-inbox/internal/config"
	"github.com/spf13/cobra"
)

func newTestCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "pr-inbox"}
	cmd.Flags().StringVarP(&repository, "repo", "R", "", "Repository in OWNER/REPO format")
	cmd.Flags().IntVarP(&prNumber, "pr", "p", 0, "Pull request number")
	cmd.Flags().StringVarP(&format, "format", "f", "md", "Output format: md or json")
	cmd.Flags().BoolVar(&includeAll, "all", false, "Include resolved threads as well")
	cmd.Flags().BoolVar(&onlyP0, "p0", false, "Show only P0 items")
	cmd.Flags().IntVar(&budget, "budget", 0, "Limit number of threads (0 = unlimited)")
	cmd.Flags().BoolVar(&includeDiff, "include-diff", false, "Include diff context for each thread")
	cmd.Flags().BoolVar(&includeTimes, "include-times", false, "Include comment timestamps")
	cmd.Flags().BoolVar(&allComments, "all-comments", false, "Include all comments for each thread (not just first/latest)")
	cmd.Flags().BoolVar(&includeIssue, "include-issue-comments", false, "Include PR conversation (issue) comments")
	cmd.Flags().BoolVar(&noUpdateCheck, "no-update-check", false, "Disable update checks")
	cmd.Flags().StringVar(&promptFile, "prompt-file", "", "Optional prompt template file")
	cmd.Flags().StringVar(&promptInline, "prompt", "", "Inline prompt template override")
	return cmd
}

func TestApplyConfigDefaults_PrecedenceCliOverConfig(t *testing.T) {
	cmd := newTestCommand()
	if err := cmd.Flags().Set("repo", "cli/repo"); err != nil {
		t.Fatalf("set repo: %v", err)
	}
	if err := cmd.Flags().Set("no-update-check", "true"); err != nil {
		t.Fatalf("set no-update-check: %v", err)
	}

	cfg := &config.Config{
		Defaults: config.Defaults{
			Repo:          "cfg/repo",
			NoUpdateCheck: false,
		},
	}
	applyConfigDefaults(cmd, cfg, false)

	if repository != "cli/repo" {
		t.Fatalf("expected repository from CLI, got %q", repository)
	}
	if !noUpdateCheck {
		t.Fatalf("expected noUpdateCheck from CLI, got false")
	}
}

func TestApplyConfigDefaults_PrecedenceConfigOverAutoDetect(t *testing.T) {
	cmd := newTestCommand()
	cfg := &config.Config{
		Defaults: config.Defaults{
			Repo:   "cfg/repo",
			PR:     123,
			Format: "json",
		},
	}

	applyConfigDefaults(cmd, cfg, false)

	if repository != "cfg/repo" {
		t.Fatalf("expected repository from config, got %q", repository)
	}
	if prNumber != 123 {
		t.Fatalf("expected prNumber from config, got %d", prNumber)
	}
	if format != "json" {
		t.Fatalf("expected format from config, got %q", format)
	}
}

func TestApplyConfigDefaults_DoesNotOverridePrFromArg(t *testing.T) {
	cmd := newTestCommand()
	prNumber = 999

	cfg := &config.Config{
		Defaults: config.Defaults{
			PR: 123,
		},
	}

	applyConfigDefaults(cmd, cfg, true)

	if prNumber != 999 {
		t.Fatalf("expected prNumber from arg to win, got %d", prNumber)
	}
}
