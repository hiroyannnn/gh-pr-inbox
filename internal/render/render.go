package render

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/hiroyannnn/gh-pr-inbox/internal/model"
)

// Markdown renders inbox items as markdown.
func Markdown(meta *model.PRMeta, items []model.InboxItem) string {
	builder := &strings.Builder{}
	summary := buildSummary(items)
	fmt.Fprintf(builder, "# PR Inbox for %s #%d\n\n", meta.Repo, meta.Number)
	fmt.Fprintf(builder, "[%s](%s)\n\n", meta.Title, meta.URL)
	if meta.Goal != "" {
		fmt.Fprintf(builder, "> %s\n\n", truncate(meta.Goal, 220))
	}

	fmt.Fprintf(builder, "Summary: P0 %d | P1 %d | P2 %d\n\n", summary.P0, summary.P1, summary.P2)
	if len(summary.HotFiles) > 0 {
		fmt.Fprintf(builder, "Hot files: %s\n\n", strings.Join(summary.HotFiles, ", "))
	}

	grouped := groupByFile(items)
	files := make([]string, 0, len(grouped))
	for file := range grouped {
		files = append(files, file)
	}
	sort.Strings(files)

	for _, file := range files {
		fmt.Fprintf(builder, "## %s\n\n", file)
		for _, item := range grouped[file] {
			fmt.Fprintf(builder, "- [%s] L%d by %s — %s\n", item.Priority, item.LineNumber, item.Author, item.Summary)
			fmt.Fprintf(builder, "  - Latest: %s\n", item.Latest)
			if item.RootCreatedAt != "" {
				fmt.Fprintf(builder, "  - Created: %s\n", item.RootCreatedAt)
			}
			if item.LatestCreatedAt != "" {
				fmt.Fprintf(builder, "  - Updated: %s\n", item.LatestCreatedAt)
			}
			fmt.Fprintf(builder, "  - Link: %s\n", item.URL)

			if item.DiffHunk != "" {
				fmt.Fprintf(builder, "  - Diff:\n\n")
				fmt.Fprintf(builder, "    ```diff\n")
				for _, line := range strings.Split(strings.TrimRight(item.DiffHunk, "\n"), "\n") {
					fmt.Fprintf(builder, "    %s\n", line)
				}
				fmt.Fprintf(builder, "    ```\n")
			}

			if len(item.Comments) > 0 {
				fmt.Fprintf(builder, "  - Comments (%d):\n", len(item.Comments))
				for _, c := range item.Comments {
					body := truncate(c.Body, 220)
					if c.CreatedAt != "" {
						fmt.Fprintf(builder, "    - %s (%s): %s\n", c.Author, c.CreatedAt, body)
					} else {
						fmt.Fprintf(builder, "    - %s: %s\n", c.Author, body)
					}
					if c.URL != "" && c.URL != item.URL {
						fmt.Fprintf(builder, "      - %s\n", c.URL)
					}
				}
			}
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

// JSON renders inbox items as JSON with PR metadata.
func JSON(meta *model.PRMeta, items []model.InboxItem) (string, error) {
	payload := map[string]any{
		"repo":   meta.Repo,
		"number": meta.Number,
		"title":  meta.Title,
		"url":    meta.URL,
		"goal":   meta.Goal,
		"items":  items,
	}
	out, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

type summaryCounts struct {
	P0       int
	P1       int
	P2       int
	HotFiles []string
}

func buildSummary(items []model.InboxItem) summaryCounts {
	counts := summaryCounts{}
	fileCounts := make(map[string]int)
	for _, item := range items {
		switch item.Priority {
		case "P0":
			counts.P0++
		case "P1":
			counts.P1++
		case "P2":
			counts.P2++
		}
		fileCounts[item.FilePath]++
	}

	type kv struct {
		File  string
		Count int
	}
	var files []kv
	for file, count := range fileCounts {
		files = append(files, kv{File: file, Count: count})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Count > files[j].Count
	})
	for i := 0; i < len(files) && i < 3; i++ {
		counts.HotFiles = append(counts.HotFiles, fmt.Sprintf("%s (%d)", files[i].File, files[i].Count))
	}

	return counts
}

func groupByFile(items []model.InboxItem) map[string][]model.InboxItem {
	grouped := make(map[string][]model.InboxItem)
	for _, item := range items {
		grouped[item.FilePath] = append(grouped[item.FilePath], item)
	}
	for _, list := range grouped {
		sort.Slice(list, func(i, j int) bool {
			if list[i].Priority != list[j].Priority {
				return list[i].Priority < list[j].Priority
			}
			return list[i].LineNumber < list[j].LineNumber
		})
	}
	return grouped
}

func truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit-3] + "..."
}
