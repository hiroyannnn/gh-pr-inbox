package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/hiroyannnn/gh-pr-inbox/internal/model"
)

var execGH = func(args ...string) ([]byte, error) {
	cmd := exec.Command("gh", args...)
	return cmd.CombinedOutput()
}

// Client handles GitHub API interactions via the gh CLI.
type Client struct {
	repository string
	owner      string
	name       string
}

// NewClient creates a new GitHub client for the given owner/repo string.
func NewClient(repository string) (*Client, error) {
	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format: %s", repository)
	}
	return &Client{repository: repository, owner: parts[0], name: parts[1]}, nil
}

// GetPRMeta fetches minimal PR metadata for context.
func (c *Client) GetPRMeta(prNumber int) (*model.PRMeta, error) {
	query := `query($owner:String!, $name:String!, $number:Int!) {
repository(owner:$owner, name:$name) {
pullRequest(number:$number) {
number
title
url
bodyText
}
}
}`

	variables := map[string]any{"owner": c.owner, "name": c.name, "number": prNumber}
	resp, err := c.runGraphQL(query, variables)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Data struct {
			Repository struct {
				PullRequest struct {
					Number   int    `json:"number"`
					Title    string `json:"title"`
					URL      string `json:"url"`
					BodyText string `json:"bodyText"`
				} `json:"pullRequest"`
			} `json:"repository"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse PR data: %w", err)
	}

	pr := parsed.Data.Repository.PullRequest
	goal := pr.BodyText
	if len(goal) > 400 {
		goal = goal[:400]
	}

	return &model.PRMeta{
		Number: pr.Number,
		Title:  pr.Title,
		URL:    pr.URL,
		Goal:   goal,
		Repo:   c.repository,
	}, nil
}

// GetReviewThreads fetches review threads for the PR using GraphQL.
func (c *Client) GetReviewThreads(prNumber int) ([]model.Thread, error) {
	var threads []model.Thread
	var cursor *string

query := `query($owner:String!, $name:String!, $number:Int!, $after:String) {
repository(owner:$owner, name:$name) {
pullRequest(number:$number) {
reviewThreads(first:100, after:$after) {
nodes {
id
isResolved
path
line
originalLine
comments(first:50) {
nodes {
id
databaseId
body
author { login }
createdAt
url
diffHunk
}
}
}
pageInfo { hasNextPage endCursor }
}
}
}
}`

	for {
		variables := map[string]any{"owner": c.owner, "name": c.name, "number": prNumber}
		if cursor != nil {
			variables["after"] = *cursor
		}

		resp, err := c.runGraphQL(query, variables)
		if err != nil {
			return nil, err
		}

		var parsed struct {
			Data struct {
				Repository struct {
					PullRequest struct {
						ReviewThreads struct {
							Nodes []struct {
								ID           string `json:"id"`
								IsResolved   bool   `json:"isResolved"`
								Path         string `json:"path"`
								Line         int    `json:"line"`
								OriginalLine int    `json:"originalLine"`
								Comments     struct {
									Nodes []struct {
										ID         string `json:"id"`
										DatabaseID int64  `json:"databaseId"`
										Body       string `json:"body"`
										Author     struct {
											Login string `json:"login"`
										} `json:"author"`
										CreatedAt string `json:"createdAt"`
										URL       string `json:"url"`
										DiffHunk   string `json:"diffHunk"`
									} `json:"nodes"`
								} `json:"comments"`
							} `json:"nodes"`
							PageInfo struct {
								HasNextPage bool   `json:"hasNextPage"`
								EndCursor   string `json:"endCursor"`
							} `json:"pageInfo"`
						} `json:"reviewThreads"`
					} `json:"pullRequest"`
				} `json:"repository"`
			} `json:"data"`
		}

		if err := json.Unmarshal(resp, &parsed); err != nil {
			return nil, fmt.Errorf("failed to parse review threads: %w", err)
		}

		rt := parsed.Data.Repository.PullRequest.ReviewThreads
		for _, node := range rt.Nodes {
			thread := model.Thread{
				ID:       node.ID,
				FilePath: node.Path,
				Line:     firstNonZero(node.Line, node.OriginalLine),
				Resolved: node.IsResolved,
			}
			if len(node.Comments.Nodes) > 0 {
				thread.URL = node.Comments.Nodes[0].URL
				thread.DiffHunk = node.Comments.Nodes[0].DiffHunk
			}
			for _, cmt := range node.Comments.Nodes {
				thread.Comments = append(thread.Comments, model.Comment{
					ID:        fmt.Sprintf("%d", cmt.DatabaseID),
					Body:      cmt.Body,
					Author:    cmt.Author.Login,
					CreatedAt: cmt.CreatedAt,
					URL:       cmt.URL,
				})
			}
			threads = append(threads, thread)
		}

		if rt.PageInfo.HasNextPage {
			cursor = &rt.PageInfo.EndCursor
			continue
		}
		break
	}

	return threads, nil
}

func (c *Client) runGraphQL(query string, variables map[string]any) ([]byte, error) {
	args := []string{"api", "graphql", "-f", fmt.Sprintf("query=%s", query)}

	keys := make([]string, 0, len(variables))
	for key := range variables {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := variables[key]
		args = append(args, "-F", fmt.Sprintf("%s=%v", key, value))
	}

	output, err := execGH(args...)
	if err != nil {
		return nil, fmt.Errorf("gh graphql failed: %w\nOutput: %s", err, string(output))
	}
	return output, nil
}

func firstNonZero(values ...int) int {
	for _, v := range values {
		if v != 0 {
			return v
		}
	}
	return 0
}
