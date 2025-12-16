# gh-pr-inbox

A GitHub CLI extension that collects PR review comments, groups them into threads, filters unresolved issues, prioritizes them, and outputs a clean, actionable "Inbox" for the PR author.

## Purpose

Reduce review noise and make it easy to decide what to fix next in your pull requests.

## Installation

```bash
gh extension install hiroyannnn/gh-pr-inbox
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

This pushes the `v1.0.0` tag and triggers `.github/workflows/release.yml` to build release assets.

### Useful commands

```bash
make help
make ci
make reinstall-local
```
