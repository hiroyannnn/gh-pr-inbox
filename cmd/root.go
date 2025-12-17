package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/hiroyannnn/gh-pr-inbox/internal/buildinfo"
	"github.com/hiroyannnn/gh-pr-inbox/internal/compact"
	"github.com/hiroyannnn/gh-pr-inbox/internal/config"
	"github.com/hiroyannnn/gh-pr-inbox/internal/github"
	"github.com/hiroyannnn/gh-pr-inbox/internal/model"
	"github.com/hiroyannnn/gh-pr-inbox/internal/render"
	"github.com/hiroyannnn/gh-pr-inbox/internal/template"
	"github.com/hiroyannnn/gh-pr-inbox/internal/updatecheck"
	"github.com/spf13/cobra"
)

var (
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
)

var rootCmd = &cobra.Command{
	Use:   "pr-inbox [PR_NUMBER]",
	Short: "Collect and organize PR review comments into an actionable inbox",
	Long: `gh pr-inbox collects PR review comments, groups them into threads,
filters unresolved issues, prioritizes them, and outputs a compact,
actionable inbox for the PR author.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInbox,
}

// Execute runs the CLI.
func Execute() error { return rootCmd.Execute() }

func init() {
	rootCmd.Flags().StringVarP(&repository, "repo", "R", "", "Repository in OWNER/REPO format")
	rootCmd.Flags().IntVarP(&prNumber, "pr", "p", 0, "Pull request number")
	rootCmd.Flags().StringVarP(&format, "format", "f", "md", "Output format: md or json")
	rootCmd.Flags().BoolVar(&includeAll, "all", false, "Include resolved threads as well")
	rootCmd.Flags().BoolVar(&onlyP0, "p0", false, "Show only P0 items")
	rootCmd.Flags().IntVar(&budget, "budget", 0, "Limit number of threads (0 = unlimited)")
	rootCmd.Flags().BoolVar(&includeDiff, "include-diff", false, "Include diff context for each thread")
	rootCmd.Flags().BoolVar(&includeTimes, "include-times", false, "Include comment timestamps")
	rootCmd.Flags().BoolVar(&allComments, "all-comments", false, "Include all comments for each thread (not just first/latest)")
	rootCmd.Flags().BoolVar(&includeIssue, "include-issue-comments", false, "Include PR conversation (issue) comments")
	rootCmd.Flags().BoolVar(&noUpdateCheck, "no-update-check", false, "Disable update checks")
	rootCmd.Flags().StringVar(&promptFile, "prompt-file", "", "Optional prompt template file")
	rootCmd.Flags().StringVar(&promptInline, "prompt", "", "Inline prompt template override")
}

func runInbox(cmd *cobra.Command, args []string) error {
	repoRoot, err := os.Getwd()
	if err != nil {
		repoRoot = ""
	}
	cfg, err := config.Load(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var updateCh <-chan string
	if !noUpdateCheck {
		// Start this early so it can run in the background while we fetch PR data.
		// We only read it later with TryReceive to avoid delaying output (especially JSON).
		updateCh = updatecheck.Start(buildinfo.Version)
	}

	prFromArg := false
	if len(args) > 0 {
		var err error
		prNumber, err = strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid PR number '%s': must be a valid integer", args[0])
		}
		prFromArg = true
	}

	applyConfigDefaults(cmd, cfg, prFromArg)

	if prNumber == 0 {
		var err error
		prNumber, err = getCurrentPRNumber()
		if err != nil || prNumber == 0 {
			return fmt.Errorf("PR number required: specify with argument, --pr flag, or run from a PR branch")
		}
	}

	if repository == "" {
		var err error
		repository, err = getCurrentRepository()
		if err != nil {
			return fmt.Errorf("repository required: specify with --repo flag or run from a git repository")
		}
	}

	client, err := github.NewClient(repository)
	if err != nil {
		return err
	}

	meta, err := client.GetPRMeta(prNumber)
	if err != nil {
		return fmt.Errorf("failed to fetch PR metadata: %w", err)
	}

	threads, err := client.GetReviewThreads(prNumber)
	if err != nil {
		return fmt.Errorf("failed to fetch review threads: %w", err)
	}

	if includeIssue {
		issueThreads, err := client.GetIssueCommentThreads(prNumber)
		if err != nil {
			return fmt.Errorf("failed to fetch issue comments: %w", err)
		}
		threads = append(threads, issueThreads...)
	}

	compactor := compact.New(compact.Options{
		IncludeResolved: includeAll,
		PriorityOnly:    priorityFilter(),
		IncludeDiff:     includeDiff,
		IncludeTimes:    includeTimes,
		AllComments:     allComments,
	})
	items := compactor.Compact(threads)
	if budget > 0 && len(items) > budget {
		items = items[:budget]
	}

	if updateCh != nil {
		if msg := updatecheck.TryReceive(updateCh); msg != "" {
			fmt.Fprintln(os.Stderr, msg)
		}
	}

	switch format {
	case "json":
		out, err := render.JSON(meta, items)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	default:
		md := render.Markdown(meta, items)
		prompt, err := resolvePrompt(cfg, md, items, meta)
		if err != nil {
			return err
		}
		fmt.Print(prompt)
		return nil
	}
}

func applyConfigDefaults(cmd *cobra.Command, cfg *config.Config, prFromArg bool) {
	flags := cmd.Flags()
	d := cfg.Defaults

	if !prFromArg && !flags.Changed("pr") && prNumber == 0 && d.PR != 0 {
		prNumber = d.PR
	}
	if !flags.Changed("repo") && repository == "" && d.Repo != "" {
		repository = d.Repo
	}
	if !flags.Changed("format") && d.Format != "" {
		format = d.Format
	}
	if !flags.Changed("all") {
		includeAll = d.All
	}
	if !flags.Changed("p0") {
		onlyP0 = d.P0
	}
	if !flags.Changed("budget") {
		budget = d.Budget
	}
	if !flags.Changed("include-diff") {
		includeDiff = d.IncludeDiff
	}
	if !flags.Changed("include-times") {
		includeTimes = d.IncludeTimes
	}
	if !flags.Changed("all-comments") {
		allComments = d.AllComments
	}
	if !flags.Changed("include-issue-comments") {
		includeIssue = d.IncludeIssueComment
	}
	if !flags.Changed("no-update-check") {
		noUpdateCheck = d.NoUpdateCheck
	}
	if !flags.Changed("prompt-file") && promptFile == "" && cfg.PromptFile != "" {
		promptFile = cfg.PromptFile
	}
}

func resolvePrompt(cfg *config.Config, md string, items []model.InboxItem, meta *model.PRMeta) (string, error) {
	prompt := cfg.Prompt
	if promptFile != "" {
		data, err := os.ReadFile(promptFile)
		if err != nil {
			return "", fmt.Errorf("failed to read prompt file %q: %w", promptFile, err)
		}
		prompt = string(data)
	}
	if promptInline != "" {
		prompt = promptInline
	}
	if prompt == "" {
		return md, nil
	}

	vars := map[string]string{
		"REPO":         meta.Repo,
		"PR_NUMBER":    strconv.Itoa(meta.Number),
		"PR_TITLE":     meta.Title,
		"PR_URL":       meta.URL,
		"PR_GOAL":      meta.Goal,
		"THREADS_MD":   md,
		"THREADS_JSON": marshalItems(items),
	}
	return template.Apply(prompt, vars), nil
}

func marshalItems(items []model.InboxItem) string {
	out, _ := json.Marshal(items)
	return string(out)
}

func priorityFilter() string {
	if onlyP0 {
		return "P0"
	}
	return ""
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
	prNum, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, fmt.Errorf("failed to parse PR number: %w", err)
	}
	return prNum, nil
}
