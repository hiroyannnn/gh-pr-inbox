# gh-pr-inbox

A GitHub CLI extension that collects PR review comments, groups them into threads, filters unresolved issues, prioritizes them, and outputs a clean, actionable "Inbox" for the PR author.

## Purpose

Reduce review noise and make it easy to decide what to fix next in your pull requests.

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

## Configuration

You can set defaults for any CLI option via YAML config files (CLI flags always override config).

Config files are loaded in this order (later files override earlier ones):

1. `~/.config/gh/pr-inbox.yml`
2. `.github/pr-inbox.yml` (in the current repository)

Example:

```yaml
defaults:
  repo: owner/repo
  pr: 123
  format: md
  all: false
  p0: false
  budget: 0
  include_diff: false
  include_times: false
  all_comments: false
  include_issue_comments: false
  no_update_check: false

prompt: |
  {{THREADS_MD}}

# Optional: default prompt file path (can still be overridden with --prompt-file)
prompt_file: ~/.config/gh/pr-inbox-prompt.txt
```

### Update notice

By default, `gh pr-inbox` checks for a newer release asynchronously and prints an upgrade hint to stderr when available (it does not block the main command output).

```bash
gh pr-inbox --no-update-check
```

### More details (diff/timestamps/all comments)
```bash
gh pr-inbox --include-diff --include-times --all-comments
```

### Include PR conversation comments
```bash
gh pr-inbox --include-issue-comments
```

## Features

- **Thread Grouping**: Automatically groups review comments into conversation threads
- **Unresolved Filtering**: Only shows comments that haven't been marked as resolved/fixed/done
- **Smart Prioritization**: Categorizes comments as HIGH 🔴, MEDIUM 🟡, or LOW 🟢 based on keywords
  - HIGH: bug, error, security, critical, blocking, broken
  - LOW: nit, minor, suggestion, optional, style
  - MEDIUM: everything else
- **Clean Output**: Displays an organized, easy-to-scan inbox view
- **JSON Export**: Machine-readable output for integration with other tools

## Example Output

```
📬 PR Review Inbox (3 items)
============================================================

[1] 🔴 HIGH
    Thread: 123456789
    File: src/main.go:45
    Author: reviewer1
    Comment: This has a potential security vulnerability - we should sanitize...
    Thread has 2 unresolved comments

[2] 🟡 MEDIUM
    Thread: 123456790
    File: cmd/root.go:120
    Author: reviewer2
    Comment: This logic seems complex, can we simplify?

[3] 🟢 LOW
    Thread: 123456791
    File: internal/util.go:30
    Author: reviewer3
    Comment: nit: consider using a more descriptive variable name
```

## How It Works

1. Fetches PR details and all review comments using GitHub CLI (`gh`)
2. Groups comments into threads based on reply relationships
3. Filters out threads that contain resolution indicators (✅, "resolved", "fixed", "done", etc.)
4. Prioritizes remaining threads based on content analysis
5. Displays results in an organized, actionable format

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
