package updatecheck

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestParseSemverTag(t *testing.T) {
	v, ok := parseSemverTag("v1.2.3")
	if !ok {
		t.Fatalf("expected ok")
	}
	if v.major != 1 || v.minor != 2 || v.patch != 3 {
		t.Fatalf("unexpected semver: %+v", v)
	}

	v, ok = parseSemverTag("v1.2.3-alpha")
	if !ok {
		t.Fatalf("expected ok for prerelease")
	}
	if v.major != 1 || v.minor != 2 || v.patch != 3 {
		t.Fatalf("unexpected semver for prerelease: %+v", v)
	}

	v, ok = parseSemverTag("v1.2.3+build123")
	if !ok {
		t.Fatalf("expected ok for build metadata")
	}
	if v.major != 1 || v.minor != 2 || v.patch != 3 {
		t.Fatalf("unexpected semver for build metadata: %+v", v)
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
	if !isNewer("v2.0.0", "v1.9.9") {
		t.Fatalf("expected newer major")
	}
	if !isNewer("v1.3.0", "v1.2.9") {
		t.Fatalf("expected newer minor")
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

func TestCheckOnce_SkipsWhenLatestTagUnparseable(t *testing.T) {
	orig := execGH
	t.Cleanup(func() { execGH = orig })

	execGH = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte("not-a-tag\n"), nil
	}

	msg, err := checkOnce("v0.1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "" {
		t.Fatalf("expected empty message, got %q", msg)
	}
}

func TestCheckOnce_ReturnsErrorWhenGHAPIFails(t *testing.T) {
	orig := execGH
	t.Cleanup(func() { execGH = orig })

	execGH = func(ctx context.Context, args ...string) ([]byte, error) {
		return nil, errors.New("boom")
	}

	_, err := checkOnce("v0.1.0")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestCheckOnce_SkipsWhenUpToDateOrNewer(t *testing.T) {
	orig := execGH
	t.Cleanup(func() { execGH = orig })

	execGH = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte("v1.2.3\n"), nil
	}

	msg, err := checkOnce("v1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "" {
		t.Fatalf("expected empty message, got %q", msg)
	}

	msg, err = checkOnce("v2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "" {
		t.Fatalf("expected empty message, got %q", msg)
	}
}

func TestStartAndTryReceive_IsAsyncAndNonBlocking(t *testing.T) {
	orig := execGH
	t.Cleanup(func() { execGH = orig })

	block := make(chan struct{})
	execGH = func(ctx context.Context, args ...string) ([]byte, error) {
		<-block
		return []byte("v9.9.9\n"), nil
	}

	ch := Start("v0.1.0")
	if msg := TryReceive(ch); msg != "" {
		t.Fatalf("expected no message yet, got %q", msg)
	}

	close(block)

	select {
	case msg := <-ch:
		if msg == "" {
			t.Fatalf("expected message")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("timed out waiting for async result")
	}
}
