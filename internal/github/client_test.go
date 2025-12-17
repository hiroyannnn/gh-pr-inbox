package github

import (
	"strings"
	"testing"
)

func TestClient_GetPRMeta_PassesVariablesAsFields(t *testing.T) {
	original := execGH
	t.Cleanup(func() { execGH = original })

	var gotArgs []string
	execGH = func(args ...string) ([]byte, error) {
		gotArgs = append([]string(nil), args...)
		return []byte(`{"data":{"repository":{"pullRequest":{"number":123,"title":"t","url":"u","bodyText":"b"}}}}`), nil
	}

	client, err := NewClient("octo/repo")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if _, err := client.GetPRMeta(123); err != nil {
		t.Fatalf("GetPRMeta: %v", err)
	}

	assertHasArg(t, gotArgs, "api")
	assertHasArg(t, gotArgs, "graphql")
	assertHasArg(t, gotArgs, "-F")
	assertHasArg(t, gotArgs, "owner=octo")
	assertHasArg(t, gotArgs, "name=repo")
	assertHasArg(t, gotArgs, "number=123")

	for _, a := range gotArgs {
		if strings.HasPrefix(a, "variables=") {
			t.Fatalf("unexpected variables= arg: %q", a)
		}
	}
}

func TestClient_GetReviewThreads_IncludesAfterCursorOnSecondPage(t *testing.T) {
	original := execGH
	t.Cleanup(func() { execGH = original })

	var calls [][]string
	var gotQuery string
	execGH = func(args ...string) ([]byte, error) {
		calls = append(calls, append([]string(nil), args...))
		if gotQuery == "" {
			gotQuery = queryArg(args)
		}
		if len(calls) == 1 {
			return []byte(`{"data":{"repository":{"pullRequest":{"reviewThreads":{"nodes":[],"pageInfo":{"hasNextPage":true,"endCursor":"CUR1"}}}}}}`), nil
		}
		return []byte(`{"data":{"repository":{"pullRequest":{"reviewThreads":{"nodes":[],"pageInfo":{"hasNextPage":false,"endCursor":""}}}}}}`), nil
	}

	client, err := NewClient("octo/repo")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if _, err := client.GetReviewThreads(123); err != nil {
		t.Fatalf("GetReviewThreads: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 GraphQL calls, got %d", len(calls))
	}

	if hasArg(calls[0], "after=CUR1") {
		t.Fatalf("did not expect after=CUR1 on first call")
	}
	if !hasArg(calls[1], "after=CUR1") {
		t.Fatalf("expected after=CUR1 on second call")
	}

	if gotQuery == "" {
		t.Fatalf("expected query arg to be passed")
	}
	if strings.Count(gotQuery, "diffHunk") != 1 {
		t.Fatalf("expected diffHunk to appear once in query, got %d", strings.Count(gotQuery, "diffHunk"))
	}
}

func TestClient_GetIssueCommentThreads_IncludesAfterCursorOnSecondPage(t *testing.T) {
	original := execGH
	t.Cleanup(func() { execGH = original })

	var calls [][]string
	var gotQuery string
	execGH = func(args ...string) ([]byte, error) {
		calls = append(calls, append([]string(nil), args...))
		if gotQuery == "" {
			gotQuery = queryArg(args)
		}
		if len(calls) == 1 {
			return []byte(`{"data":{"repository":{"pullRequest":{"comments":{"nodes":[],"pageInfo":{"hasNextPage":true,"endCursor":"CUR1"}}}}}}`), nil
		}
		return []byte(`{"data":{"repository":{"pullRequest":{"comments":{"nodes":[],"pageInfo":{"hasNextPage":false,"endCursor":""}}}}}}`), nil
	}

	client, err := NewClient("octo/repo")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if _, err := client.GetIssueCommentThreads(123); err != nil {
		t.Fatalf("GetIssueCommentThreads: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 GraphQL calls, got %d", len(calls))
	}

	if hasArg(calls[0], "after=CUR1") {
		t.Fatalf("did not expect after=CUR1 on first call")
	}
	if !hasArg(calls[1], "after=CUR1") {
		t.Fatalf("expected after=CUR1 on second call")
	}

	if gotQuery == "" {
		t.Fatalf("expected query arg to be passed")
	}
	if !strings.Contains(gotQuery, "comments(first:100") {
		t.Fatalf("expected comments query, got: %s", gotQuery)
	}
}

func assertHasArg(t *testing.T, args []string, want string) {
	t.Helper()
	if !hasArg(args, want) {
		t.Fatalf("missing arg %q in %v", want, args)
	}
}

func hasArg(args []string, want string) bool {
	for _, a := range args {
		if a == want {
			return true
		}
	}
	return false
}

func queryArg(args []string) string {
	for _, a := range args {
		if strings.HasPrefix(a, "query=") {
			return a
		}
	}
	return ""
}
