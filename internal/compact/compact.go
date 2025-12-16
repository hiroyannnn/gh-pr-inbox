package compact

import (
	"sort"
	"strings"

	"github.com/hiroyannnn/gh-pr-inbox/internal/model"
)

// Options controls compaction behavior.
type Options struct {
	IncludeResolved bool
	PriorityOnly    string
	ReplyBudget     int
}

// Compactor transforms threads into prioritized inbox items.
type Compactor struct {
	options Options
}

// New creates a Compactor with the given options.
func New(options Options) *Compactor {
	return &Compactor{options: options}
}

// Compact converts threads into inbox items applying filtering and prioritization.
func (c *Compactor) Compact(threads []model.Thread) []model.InboxItem {
	var items []model.InboxItem

	for _, thread := range threads {
		if !c.options.IncludeResolved && thread.Resolved {
			continue
		}

		if len(thread.Comments) == 0 {
			continue
		}

		root := thread.Comments[0]
		latest := thread.Comments[len(thread.Comments)-1]

		item := model.InboxItem{
			ThreadID:   thread.ID,
			Priority:   determinePriority(thread, root.Body),
			FilePath:   thread.FilePath,
			LineNumber: thread.Line,
			Author:     root.Author,
			Summary:    condense(root.Body),
			Latest:     condense(latest.Body),
			URL:        chooseURL(thread, latest),
			Resolved:   thread.Resolved,
		}

		if c.options.PriorityOnly != "" && item.Priority != c.options.PriorityOnly {
			continue
		}

		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Priority != items[j].Priority {
			return priorityRank(items[i].Priority) < priorityRank(items[j].Priority)
		}
		return items[i].FilePath < items[j].FilePath
	})

	return items
}

func determinePriority(thread model.Thread, body string) string {
	text := strings.ToLower(body)

	highSignals := []string{"must", "block", "blocking", "security", "crash", "bug", "failure", "incorrect"}
	for _, sig := range highSignals {
		if strings.Contains(text, sig) {
			return "P0"
		}
	}

	if len(thread.Comments) > 4 {
		return "P1"
	}

	lowSignals := []string{"nit", "nitpick", "style", "optional", "suggest", "tiny"}
	for _, sig := range lowSignals {
		if strings.Contains(text, sig) {
			return "P2"
		}
	}

	return "P1"
}

func condense(body string) string {
	body = strings.TrimSpace(body)
	if len(body) > 220 {
		return body[:217] + "..."
	}
	return body
}

func chooseURL(thread model.Thread, latest model.Comment) string {
	if latest.URL != "" {
		return latest.URL
	}
	return thread.URL
}

func clamp(val, min, max int) int {
	if val == 0 {
		return max
	}
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func priorityRank(p string) int {
	switch p {
	case "P0":
		return 0
	case "P1":
		return 1
	case "P2":
		return 2
	default:
		return 3
	}
}
