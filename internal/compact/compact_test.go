package compact

import (
	"strings"
	"testing"

	"github.com/hiroyannnn/gh-pr-inbox/internal/model"
)

func TestCompact_FiltersResolvedAndAssignsPriority(t *testing.T) {
	threads := []model.Thread{
		{
			ID:       "t1",
			FilePath: "foo.go",
			Line:     12,
			Resolved: false,
			Comments: []model.Comment{
				{Body: "must handle crash", Author: "alice", URL: "u1"},
				{Body: "please add test", Author: "bob", URL: "u2"},
			},
		},
		{
			ID:       "t2",
			FilePath: "bar.go",
			Line:     8,
			Resolved: true,
			Comments: []model.Comment{{Body: "nit: spacing", Author: "carol", URL: "u3"}},
		},
	}

	compactor := New(Options{IncludeResolved: false})
	items := compactor.Compact(threads)

	if len(items) != 1 {
		t.Fatalf("expected only unresolved thread kept, got %d", len(items))
	}
	item := items[0]
	if item.ThreadID != "t1" {
		t.Fatalf("unexpected thread kept: %s", item.ThreadID)
	}
	if item.Priority != "P0" {
		t.Fatalf("expected P0 priority from blocking language, got %s", item.Priority)
	}
	if item.Latest != "please add test" {
		t.Fatalf("expected latest reply trimmed to budget, got %q", item.Latest)
	}
}

func TestCompact_PriorityFilteringAndOrdering(t *testing.T) {
	threads := []model.Thread{
		{
			ID:       "t1",
			FilePath: "b.go",
			Line:     10,
			Comments: []model.Comment{{Body: "nit: rename var"}},
		},
		{
			ID:       "t2",
			FilePath: "a.go",
			Line:     5,
			Comments: []model.Comment{{Body: "crash on start"}},
		},
	}

	compactor := New(Options{PriorityOnly: "P0"})
	items := compactor.Compact(threads)

	if len(items) != 1 {
		t.Fatalf("expected only high priority threads, got %d", len(items))
	}
	if items[0].ThreadID != "t2" {
		t.Fatalf("expected P0 thread first, got %s", items[0].ThreadID)
	}
	if items[0].FilePath != "a.go" {
		t.Fatalf("expected items ordered by file within priority, got %s", items[0].FilePath)
	}
}

func TestCompact_CondenseText(t *testing.T) {
	threads := []model.Thread{
		{
			ID:       "t1",
			FilePath: "foo.go",
			Line:     3,
			Comments: []model.Comment{
				{Body: strings.Repeat("a", 230)},
				{Body: "final"},
			},
		},
	}

	compactor := New(Options{})
	items := compactor.Compact(threads)

	if items[0].Summary != strings.Repeat("a", 217)+"..." {
		t.Fatalf("expected summary condensed to 220 runes, got %q", items[0].Summary)
	}
	if items[0].Latest != "final" {
		t.Fatalf("expected latest reply to be kept, got %q", items[0].Latest)
	}
}

func TestCompact_CondenseText_WithMultibyteChars(t *testing.T) {
	// Test with emojis and multi-byte characters
	longTextWithEmojis := strings.Repeat("🎉", 230) // Each emoji is multiple bytes
	threads := []model.Thread{
		{
			ID:       "t1",
			FilePath: "foo.go",
			Line:     3,
			Comments: []model.Comment{
				{Body: longTextWithEmojis},
			},
		},
	}

	compactor := New(Options{})
	items := compactor.Compact(threads)

	expected := strings.Repeat("🎉", 217) + "..."
	if items[0].Summary != expected {
		t.Fatalf("expected summary with emojis condensed to 220 runes, got %q (len=%d runes)", items[0].Summary, len([]rune(items[0].Summary)))
	}

	// Verify the result is valid UTF-8
	if !isValidUTF8(items[0].Summary) {
		t.Fatalf("condensed summary is not valid UTF-8")
	}
}

func isValidUTF8(s string) bool {
	for _, r := range s {
		if r == '\uFFFD' { // Unicode replacement character indicates invalid UTF-8
			return false
		}
	}
	return true
}
