package updatecheck

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	extensionOwner = "hiroyannnn"
	extensionRepo  = "gh-pr-inbox"
)

var execGH = func(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gh", args...)
	return cmd.CombinedOutput()
}

// Start checks for a newer GitHub Release asynchronously.
// It returns a channel that will receive at most one result then be closed.
func Start(currentVersion string) <-chan string {
	ch := make(chan string, 1)
	go func() {
		defer close(ch)

		msg, err := checkOnce(currentVersion)
		if err != nil || msg == "" {
			return
		}
		ch <- msg
	}()
	return ch
}

// TryReceive returns an update message if it's ready (non-blocking).
func TryReceive(ch <-chan string) string {
	select {
	case msg, ok := <-ch:
		if !ok {
			return ""
		}
		return msg
	default:
		return ""
	}
}

func checkOnce(currentVersion string) (string, error) {
	currentVersion = strings.TrimSpace(currentVersion)
	if currentVersion == "" || currentVersion == "dev" {
		return "", nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	latestTag, err := latestReleaseTag(ctx)
	if err != nil {
		return "", err
	}

	// Best-effort: if either tag isn't a plain semver tag (vMAJOR.MINOR.PATCH) we skip the notice.
	if ok := isNewer(latestTag, currentVersion); !ok {
		return "", nil
	}

	return fmt.Sprintf(
		"Update available: %s (current %s). Run: gh extension upgrade %s",
		latestTag,
		currentVersion,
		extensionRepo,
	), nil
}

func latestReleaseTag(ctx context.Context) (string, error) {
	endpoint := fmt.Sprintf("repos/%s/%s/releases/latest", extensionOwner, extensionRepo)
	out, err := execGH(ctx, "api", endpoint, "-q", ".tag_name")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func isNewer(latestTag, currentTag string) bool {
	latest, ok := parseSemverTag(latestTag)
	if !ok {
		return false
	}
	current, ok := parseSemverTag(currentTag)
	if !ok {
		return false
	}
	return compareSemver(latest, current) > 0
}

type semver struct {
	major int
	minor int
	patch int
}

func parseSemverTag(tag string) (semver, bool) {
	tag = strings.TrimSpace(tag)
	if !strings.HasPrefix(tag, "v") {
		return semver{}, false
	}
	tag = strings.TrimPrefix(tag, "v")
	tag = strings.SplitN(tag, "-", 2)[0]
	tag = strings.SplitN(tag, "+", 2)[0]

	parts := strings.Split(tag, ".")
	if len(parts) != 3 {
		return semver{}, false
	}

	if hasLeadingZeros(parts[0]) || hasLeadingZeros(parts[1]) || hasLeadingZeros(parts[2]) {
		return semver{}, false
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return semver{}, false
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return semver{}, false
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return semver{}, false
	}
	if major < 0 || minor < 0 || patch < 0 {
		return semver{}, false
	}
	return semver{major: major, minor: minor, patch: patch}, true
}

func hasLeadingZeros(s string) bool {
	if len(s) <= 1 {
		return false
	}
	return strings.HasPrefix(s, "0")
}

func compareSemver(a, b semver) int {
	if a.major != b.major {
		return a.major - b.major
	}
	if a.minor != b.minor {
		return a.minor - b.minor
	}
	return a.patch - b.patch
}
