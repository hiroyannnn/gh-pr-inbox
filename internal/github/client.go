package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// Client handles GitHub API interactions via gh CLI
type Client struct {
	repository string
}

// NewClient creates a new GitHub client
func NewClient(repository string) *Client {
	return &Client{
		repository: repository,
	}
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
	URL string `json:"url"`
}

// ReviewComment represents a review comment on a PR
type ReviewComment struct {
	ID                  int64  `json:"id"`
	Body                string `json:"body"`
	Path                string `json:"path"`
	Line                int    `json:"line"`
	OriginalLine        int    `json:"originalLine"`
	DiffHunk            string `json:"diffHunk"`
	Position            int    `json:"position"`
	OriginalPosition    int    `json:"originalPosition"`
	CommitID            string `json:"commitId"`
	OriginalCommitID    string `json:"originalCommitId"`
	InReplyToID         int64  `json:"inReplyToId"`
	User                User   `json:"user"`
	CreatedAt           string `json:"createdAt"`
	UpdatedAt           string `json:"updatedAt"`
	PullRequestReviewID int64  `json:"pullRequestReviewId"`
}

// User represents a GitHub user
type User struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
}

// GetPullRequest fetches PR details
func (c *Client) GetPullRequest(prNumber int) (*PullRequest, error) {
	args := []string{
		"pr", "view", fmt.Sprintf("%d", prNumber),
		"--repo", c.repository,
		"--json", "number,title,state,author,url",
	}

	output, err := c.runGHCommand(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	var pr PullRequest
	if err := json.Unmarshal(output, &pr); err != nil {
		return nil, fmt.Errorf("failed to parse PR data: %w", err)
	}

	return &pr, nil
}

// GetReviewComments fetches all review comments for a PR
func (c *Client) GetReviewComments(prNumber int) ([]ReviewComment, error) {
	// Use gh api to get review comments with more details
	endpoint := fmt.Sprintf("repos/%s/pulls/%d/comments", c.repository, prNumber)
	args := []string{
		"api",
		endpoint,
		"--paginate",
		"-q", ".",
	}

	output, err := c.runGHCommand(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get review comments: %w", err)
	}

	var comments []ReviewComment
	if err := json.Unmarshal(output, &comments); err != nil {
		return nil, fmt.Errorf("failed to parse review comments: %w", err)
	}

	return comments, nil
}

// runGHCommand executes a gh CLI command and returns the output
func (c *Client) runGHCommand(args ...string) ([]byte, error) {
	cmd := exec.Command("gh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("gh command failed: %w\nOutput: %s", err, string(output))
	}
	return output, nil
}
