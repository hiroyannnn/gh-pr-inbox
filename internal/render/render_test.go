package render

import (
	"strings"
	"testing"

	"github.com/hiroyannnn/gh-pr-inbox/internal/model"
)

func TestMarkdownIncludesSummaryAndGrouping(t *testing.T) {
	meta := &model.PRMeta{Repo: "acme/widgets", Number: 42, Title: "Add feature", URL: "http://example.com"}
	items := []model.InboxItem{
		{ThreadID: "t1", Priority: "P0", FilePath: "api/service.go", LineNumber: 12, Author: "alice", Summary: "must fix", Latest: "do it", URL: "u1"},
		{ThreadID: "t2", Priority: "P1", FilePath: "api/service.go", LineNumber: 20, Author: "bob", Summary: "consider", Latest: "ping", URL: "u2"},
		{ThreadID: "t3", Priority: "P2", FilePath: "ui/view.go", LineNumber: 3, Author: "carol", Summary: "nit", Latest: "nit", URL: "u3"},
	}

	out := Markdown(meta, items)

	required := []string{
		"Summary: P0 1 | P1 1 | P2 1",
		"Hot files: api/service.go (2)",
		"## api/service.go",
		"- [P0] L12 by alice — must fix",
		"- Latest: do it",
		"- Link: u1",
	}
	for _, expected := range required {
		if !strings.Contains(out, expected) {
			t.Fatalf("expected markdown to contain %q", expected)
		}
	}
}

func TestJSONRendering(t *testing.T) {
	meta := &model.PRMeta{Repo: "acme/widgets", Number: 1, Title: "Demo", URL: "u", Goal: "goal"}
	items := []model.InboxItem{{ThreadID: "t1", Priority: "P1"}}

	out, err := JSON(meta, items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "\"items\"") || !strings.Contains(out, "\"goal\"") {
		t.Fatalf("json output missing fields: %s", out)
	}
}
