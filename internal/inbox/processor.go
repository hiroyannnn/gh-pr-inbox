package inbox

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/hiroyannnn/gh-pr-inbox/internal/github"
)

// Priority levels for inbox items
const (
	PriorityHigh   = "🔴 HIGH"
	PriorityMedium = "🟡 MEDIUM"
	PriorityLow    = "🟢 LOW"

	// HighPriorityThreadThreshold is the number of unresolved comments that triggers high priority
	HighPriorityThreadThreshold = 3
)

// InboxItem represents a single item in the PR inbox
type InboxItem struct {
	ThreadID        string `json:"threadId"`
	Priority        string `json:"priority"`
	FilePath        string `json:"filePath"`
	LineNumber      int    `json:"lineNumber"`
	Author          string `json:"author"`
	Body            string `json:"body"`
	UnresolvedCount int    `json:"unresolvedCount"`
	CreatedAt       string `json:"createdAt"`
	URL             string `json:"url"`
}

// Processor handles the organization and prioritization of PR comments
type Processor struct{}

// NewProcessor creates a new inbox processor
func NewProcessor() *Processor {
	return &Processor{}
}

// Process takes PR data and comments and returns organized inbox items
func (p *Processor) Process(pr *github.PullRequest, comments []github.ReviewComment) []InboxItem {
	// Group comments into threads
	threads := p.groupIntoThreads(comments)

	// Filter unresolved threads
	unresolvedThreads := p.filterUnresolved(threads, comments)

	// Convert to inbox items
	items := p.createInboxItems(unresolvedThreads, pr)

	// Prioritize
	p.prioritize(items)

	// Sort by priority (high to low) and then by creation time
	sort.Slice(items, func(i, j int) bool {
		if items[i].Priority != items[j].Priority {
			return priorityValue(items[i].Priority) > priorityValue(items[j].Priority)
		}
		return items[i].CreatedAt < items[j].CreatedAt
	})

	return items
}

// groupIntoThreads groups comments by their thread relationship
func (p *Processor) groupIntoThreads(comments []github.ReviewComment) map[string][]github.ReviewComment {
	threads := make(map[string][]github.ReviewComment)

	for _, comment := range comments {
		// Determine thread ID - use the root comment ID or the comment's own ID if it's a root
		threadID := p.getThreadID(comment, comments)
		threads[threadID] = append(threads[threadID], comment)
	}

	return threads
}

// getThreadID determines the thread ID for a comment
func (p *Processor) getThreadID(comment github.ReviewComment, allComments []github.ReviewComment) string {
	// If this is a reply, find the root comment
	if comment.InReplyToID > 0 {
		for _, c := range allComments {
			if c.ID == comment.InReplyToID {
				// Recursively find the root
				return p.getThreadID(c, allComments)
			}
		}
	}

	// This is a root comment
	return fmt.Sprintf("%d", comment.ID)
}

// filterUnresolved filters threads to only include unresolved ones
func (p *Processor) filterUnresolved(threads map[string][]github.ReviewComment, allComments []github.ReviewComment) map[string][]github.ReviewComment {
	unresolved := make(map[string][]github.ReviewComment)

	for threadID, threadComments := range threads {
		if !p.isThreadResolved(threadComments) {
			unresolved[threadID] = threadComments
		}
	}

	return unresolved
}

// isThreadResolved checks if a thread appears to be resolved
func (p *Processor) isThreadResolved(threadComments []github.ReviewComment) bool {
	// Regex pattern to match "done" as a whole word
	donePattern := regexp.MustCompile(`\bdone\b`)

	// Check for resolution indicators
	for _, comment := range threadComments {
		body := strings.ToLower(comment.Body)

		// Look for common resolution phrases
		if strings.Contains(body, "✅") ||
			strings.Contains(body, "resolved") ||
			strings.Contains(body, "fixed") ||
			donePattern.MatchString(body) ||
			strings.Contains(body, "addressed") ||
			strings.Contains(body, "closing") {
			return true
		}
	}

	return false
}

// createInboxItems converts threads to inbox items
func (p *Processor) createInboxItems(threads map[string][]github.ReviewComment, pr *github.PullRequest) []InboxItem {
	var items []InboxItem

	for threadID, threadComments := range threads {
		// Use the first comment in the thread as the primary one
		if len(threadComments) == 0 {
			continue
		}

		// Sort thread comments by creation time
		sort.Slice(threadComments, func(i, j int) bool {
			return threadComments[i].CreatedAt < threadComments[j].CreatedAt
		})

		rootComment := threadComments[0]

		item := InboxItem{
			ThreadID:        threadID,
			FilePath:        rootComment.Path,
			LineNumber:      rootComment.Line,
			Author:          rootComment.User.Login,
			Body:            rootComment.Body,
			UnresolvedCount: len(threadComments),
			CreatedAt:       rootComment.CreatedAt,
			URL:             fmt.Sprintf("%s#discussion_r%d", pr.URL, rootComment.ID),
		}

		items = append(items, item)
	}

	return items
}

// prioritize assigns priority levels to inbox items
func (p *Processor) prioritize(items []InboxItem) {
	for i := range items {
		items[i].Priority = p.determinePriority(&items[i])
	}
}

// determinePriority determines the priority level of an inbox item
func (p *Processor) determinePriority(item *InboxItem) string {
	body := strings.ToLower(item.Body)

	// High priority indicators
	if strings.Contains(body, "bug") ||
		strings.Contains(body, "error") ||
		strings.Contains(body, "issue") ||
		strings.Contains(body, "problem") ||
		strings.Contains(body, "critical") ||
		strings.Contains(body, "security") ||
		strings.Contains(body, "vulnerability") ||
		strings.Contains(body, "must") ||
		strings.Contains(body, "blocking") ||
		strings.Contains(body, "broken") {
		return PriorityHigh
	}

	// Low priority indicators
	if strings.Contains(body, "nit") ||
		strings.Contains(body, "minor") ||
		strings.Contains(body, "suggestion") ||
		strings.Contains(body, "optional") ||
		strings.Contains(body, "consider") ||
		strings.Contains(body, "maybe") ||
		strings.Contains(body, "style") {
		return PriorityLow
	}

	// Multiple unresolved comments in thread increases priority
	if item.UnresolvedCount > HighPriorityThreadThreshold {
		return PriorityHigh
	}

	// Default to medium
	return PriorityMedium
}

// priorityValue returns a numeric value for sorting priorities
func priorityValue(priority string) int {
	switch priority {
	case PriorityHigh:
		return 3
	case PriorityMedium:
		return 2
	case PriorityLow:
		return 1
	default:
		return 0
	}
}
