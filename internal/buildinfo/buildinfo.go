package buildinfo

// Version is set at build time via -ldflags.
// Example: -X github.com/hiroyannnn/gh-pr-inbox/internal/buildinfo.Version=v0.1.0
var Version = "dev"
