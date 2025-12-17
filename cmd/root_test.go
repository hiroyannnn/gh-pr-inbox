package cmd

import (
	"testing"

	"github.com/hiroyannnn/gh-pr-inbox/internal/config"
	"github.com/spf13/cobra"
)

type globalsSnapshot struct {
	repository    string
	prNumber      int
	format        string
	includeAll    bool
	onlyP0        bool
	budget        int
	includeDiff   bool
	includeTimes  bool
	allComments   bool
	includeIssue  bool
	noUpdateCheck bool
	promptFile    string
	promptInline  string
}

func snapshotGlobals() globalsSnapshot {
	return globalsSnapshot{
		repository:    repository,
		prNumber:      prNumber,
		format:        format,
		includeAll:    includeAll,
		onlyP0:        onlyP0,
		budget:        budget,
		includeDiff:   includeDiff,
		includeTimes:  includeTimes,
		allComments:   allComments,
		includeIssue:  includeIssue,
		noUpdateCheck: noUpdateCheck,
		promptFile:    promptFile,
		promptInline:  promptInline,
	}
}

func (s globalsSnapshot) restore() {
	repository = s.repository
	prNumber = s.prNumber
	format = s.format
	includeAll = s.includeAll
	onlyP0 = s.onlyP0
	budget = s.budget
	includeDiff = s.includeDiff
	includeTimes = s.includeTimes
	allComments = s.allComments
	includeIssue = s.includeIssue
	noUpdateCheck = s.noUpdateCheck
	promptFile = s.promptFile
	promptInline = s.promptInline
}

func setupTestCommand(t *testing.T) *cobra.Command {
	t.Helper()
	snap := snapshotGlobals()
	t.Cleanup(snap.restore)
	return newTestCommand()
}

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
	cmd := setupTestCommand(t)
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
	cmd := setupTestCommand(t)
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
	cmd := setupTestCommand(t)
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

func TestApplyConfigDefaults_BoolDefaultsFromConfig(t *testing.T) {
	cmd := setupTestCommand(t)
	cfg := &config.Config{
		Defaults: config.Defaults{
			All:                  true,
			P0:                   true,
			IncludeDiff:          true,
			IncludeTimes:         true,
			AllComments:          true,
			IncludeIssueComments: true,
			NoUpdateCheck:        true,
		},
	}

	applyConfigDefaults(cmd, cfg, false)

	if !includeAll || !onlyP0 || !includeDiff || !includeTimes || !allComments || !includeIssue || !noUpdateCheck {
		t.Fatalf("expected bool defaults from config to apply")
	}
}

func TestApplyConfigDefaults_BoolCliWinsOverConfig(t *testing.T) {
	cmd := setupTestCommand(t)
	if err := cmd.Flags().Set("all", "false"); err != nil {
		t.Fatalf("set all: %v", err)
	}
	if err := cmd.Flags().Set("include-issue-comments", "false"); err != nil {
		t.Fatalf("set include-issue-comments: %v", err)
	}

	cfg := &config.Config{
		Defaults: config.Defaults{
			All:                  true,
			IncludeIssueComments: true,
		},
	}

	applyConfigDefaults(cmd, cfg, false)

	if includeAll {
		t.Fatalf("expected --all (explicit false) to win over config true")
	}
	if includeIssue {
		t.Fatalf("expected --include-issue-comments (explicit false) to win over config true")
	}
}
