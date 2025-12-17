package updatecheck

import (
	"context"
	"strings"
	"testing"
)

func TestParseSemverTag(t *testing.T) {
	v, ok := parseSemverTag("v1.2.3")
	if !ok {
		t.Fatalf("expected ok")
	}
	if v.major != 1 || v.minor != 2 || v.patch != 3 {
		t.Fatalf("unexpected semver: %+v", v)
	}

	if _, ok := parseSemverTag("1.2.3"); ok {
		t.Fatalf("expected non-v tag to fail")
	}
	if _, ok := parseSemverTag("v1.2"); ok {
		t.Fatalf("expected short tag to fail")
	}
}

func TestIsNewer(t *testing.T) {
	if !isNewer("v1.2.4", "v1.2.3") {
		t.Fatalf("expected newer")
	}
	if isNewer("v1.2.3", "v1.2.3") {
		t.Fatalf("expected not newer")
	}
	if isNewer("v1.2.3", "v1.2.4") {
		t.Fatalf("expected not newer")
	}
}

func TestCheckOnce_PrintsUpgradeHint(t *testing.T) {
	orig := execGH
	t.Cleanup(func() { execGH = orig })

	execGH = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte("v9.9.9\n"), nil
	}

	msg, err := checkOnce("v0.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg == "" {
		t.Fatalf("expected message")
	}
	if !strings.Contains(msg, "Update available") || !strings.Contains(msg, "gh extension upgrade") {
		t.Fatalf("unexpected message: %q", msg)
	}
}

func TestCheckOnce_SkipsDevBuild(t *testing.T) {
	orig := execGH
	t.Cleanup(func() { execGH = orig })

	called := false
	execGH = func(ctx context.Context, args ...string) ([]byte, error) {
		called = true
		return []byte("v9.9.9\n"), nil
	}

	msg, err := checkOnce("dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "" {
		t.Fatalf("expected empty message, got %q", msg)
	}
	if called {
		t.Fatalf("expected no gh call for dev")
	}
}
