# gh-pr-inbox

A GitHub CLI extension that extracts unresolved PR review comments, prioritizes them, and outputs a clean format optimized for LLM/AI agent consumption.

## Purpose

Make it easy to feed PR review comments to LLMs like Claude Code or Cursor Agent. The tool filters out resolved threads, assigns priority levels (P0/P1/P2), and outputs Markdown or JSON that's ready to use as prompt input.

## Installation

```bash
gh extension install hiroyannnn/gh-pr-inbox
```

If you want a shorter command, set an alias:

```bash
gh alias set pri 'pr-inbox'
```

Or build from source:

```bash
git clone https://github.com/hiroyannnn/gh-pr-inbox.git
cd gh-pr-inbox
go build -o gh-pr-inbox .
```

## Usage

### View inbox for current PR (from PR branch)

```bash
gh pr-inbox
gh pri
```

### View inbox for specific PR

```bash
gh pr-inbox 123
# or
gh pr-inbox --pr 123
```

### View inbox for PR in specific repository

```bash
gh pr-inbox --repo owner/repo --pr 123
```

### Output as JSON

```bash
gh pr-inbox --format json
```

### Pipe to LLM

Pipe the output to any LLM CLI tool (e.g., [Claude Code](https://docs.anthropic.com/en/docs/claude-code), [GitHub Copilot CLI](https://docs.github.com/en/copilot/github-copilot-in-the-cli), or [aichat](https://github.com/sigoden/aichat)):

```bash
# Example with Claude Code (requires: npm install -g @anthropic-ai/claude-code)
gh pr-inbox | claude "Fix these review comments"

# Example with other LLM tools
gh pr-inbox | your-llm-cli "Fix these review comments"
```

## Example Output

Default output is Markdown, grouped by file:

```markdown
# PR Inbox for owner/repo #123

[Fix authentication bug](https://github.com/owner/repo/pull/123)

> PR description text here

Summary: P0 1 | P1 2 | P2 1

Hot files: src/auth.go (2), cmd/root.go (1)

## src/auth.go

- [P0] L45 by reviewer1 — This has a potential security vulnerability
  - Latest: We should sanitize user input here
  - Link: https://github.com/owner/repo/pull/123#discussion_r123456789

- [P1] L120 by reviewer2 — This logic seems complex, can we simplify?
  - Latest: Consider extracting this into a separate function
  - Link: https://github.com/owner/repo/pull/123#discussion_r123456790

## internal/util.go

- [P2] L30 by reviewer3 — nit: consider using a more descriptive variable name
  - Latest: Maybe rename 'x' to 'userCount'?
  - Link: https://github.com/owner/repo/pull/123#discussion_r123456791
```

## Features

- **Thread Grouping**: Automatically groups review comments into conversation threads
- **Unresolved Filtering**: Only shows threads not marked as resolved via GitHub's "Resolve conversation" feature
- **Smart Prioritization**: Categorizes comments as P0, P1, or P2 based on keywords
  - P0 (high): must, block, blocking, security, crash, bug, failure, incorrect
  - P2 (low): nit, nitpick, style, optional, suggest, tiny
  - P1 (medium): everything else
  - Note: Threads with 5+ comments are promoted to P1 (indicating active discussion)
- **LLM-Ready Output**: Markdown format optimized for LLM consumption
- **JSON Export**: Machine-readable output for programmatic use
- **Prompt Templates**: Wrap the output in custom prompt text using `--prompt` or `--prompt-file`. Supports template variables like `{{THREADS_MD}}` to embed the review comments (see [Prompt Template Variables](#prompt-template-variables) for details)

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `--pr`, `-p` | PR number | Current branch's PR |
| `--repo`, `-R` | Repository (owner/repo) | Current repository |
| `--format`, `-f` | Output format (md/json) | md |
| `--all` | Include resolved threads | false |
| `--p0` | Show only P0 items | false |
| `--budget` | Limit number of threads (0 = unlimited) | 0 |
| `--include-diff` | Include diff context | false |
| `--include-times` | Include timestamps | false |
| `--all-comments` | Include all comments in thread | false |
| `--include-issue-comments` | Include PR conversation comments | false |
| `--no-update-check` | Disable update check | false |
| `--prompt-file` | Path to prompt template file | - |
| `--prompt` | Inline prompt template | - |

## Configuration

You can set defaults for any CLI option via YAML config files (CLI flags always override config).

Config files are loaded in this order (later files override earlier ones):

1. `~/.config/gh/pr-inbox.yml`
2. `.github/pr-inbox.yml` (in the current repository)

Example (customize based on your workflow):

```yaml
defaults:
  format: md           # output format (md or json)
  # include_diff: true   # uncomment to include diff context (default: false)
  # all_comments: true   # uncomment to show all comments in thread (default: false)
  no_update_check: true  # disable update check

prompt: |
  Please fix the following review comments:

  {{THREADS_MD}}

# prompt_file: ~/.config/gh/pr-inbox-prompt.txt  # alternative: load prompt from file
```

### Prompt Template Variables

| Variable | Content |
|----------|---------|
| `{{REPO}}` | Repository name (owner/repo) |
| `{{PR_NUMBER}}` | PR number |
| `{{PR_TITLE}}` | PR title |
| `{{PR_URL}}` | PR URL |
| `{{PR_GOAL}}` | PR description |
| `{{THREADS_MD}}` | Markdown formatted threads |
| `{{THREADS_JSON}}` | JSON formatted threads |

## How It Works

1. Fetches PR details and all review comments using GitHub CLI (`gh`)
2. Groups comments into threads based on reply relationships
3. Filters out threads marked as resolved via GitHub's "Resolve conversation"
4. Prioritizes remaining threads based on keyword analysis
5. Outputs results in Markdown or JSON format

## Requirements

- [GitHub CLI](https://cli.github.com/) (`gh`) installed and authenticated
- Go 1.21 or higher (for building from source)

## License

MIT

## Development

### Release

```bash
make release version=1.0.0
```

This pushes the `v1.0.0` tag and triggers the release workflow (`.github/workflows/release.yml`) to build release assets.

### Useful commands

```bash
make help
make ci
make reinstall-local
```
