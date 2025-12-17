package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestConfigUnmarshalDefaultsAndPrompt(t *testing.T) {
	data := []byte(`
prompt: "hello {{PR_TITLE}}"
prompt_file: "/tmp/prompt.txt"
defaults:
  repo: "octo/repo"
  pr: 123
  format: "json"
  all: true
  p0: true
  budget: 10
  include_diff: true
  include_times: true
  all_comments: true
  include_issue_comments: true
  no_update_check: true
`)

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if cfg.Prompt == "" || cfg.PromptFile == "" {
		t.Fatalf("expected prompt and prompt_file to be set")
	}
	if cfg.Defaults.Repo != "octo/repo" || cfg.Defaults.PR != 123 || cfg.Defaults.Format != "json" {
		t.Fatalf("unexpected defaults: %+v", cfg.Defaults)
	}
	if !cfg.Defaults.All || !cfg.Defaults.P0 || cfg.Defaults.Budget != 10 {
		t.Fatalf("unexpected defaults: %+v", cfg.Defaults)
	}
	if !cfg.Defaults.IncludeDiff || !cfg.Defaults.IncludeTimes || !cfg.Defaults.AllComments {
		t.Fatalf("unexpected defaults: %+v", cfg.Defaults)
	}
	if !cfg.Defaults.IncludeIssueComments || !cfg.Defaults.NoUpdateCheck {
		t.Fatalf("unexpected defaults: %+v", cfg.Defaults)
	}
}

func TestLoad_MergesGlobalThenRepoOverrides(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	globalPath := filepath.Join(home, ".config", "gh", "pr-inbox.yml")
	if err := os.MkdirAll(filepath.Dir(globalPath), 0o755); err != nil {
		t.Fatalf("mkdir global config dir: %v", err)
	}
	if err := os.WriteFile(globalPath, []byte(`
defaults:
  repo: "global/repo"
  pr: 111
  include_diff: true
`), 0o644); err != nil {
		t.Fatalf("write global config: %v", err)
	}

	repoRoot := t.TempDir()
	repoPath := filepath.Join(repoRoot, ".github", "pr-inbox.yml")
	if err := os.MkdirAll(filepath.Dir(repoPath), 0o755); err != nil {
		t.Fatalf("mkdir repo config dir: %v", err)
	}
	if err := os.WriteFile(repoPath, []byte(`
defaults:
  repo: "repo/repo"
  pr: 222
`), 0o644); err != nil {
		t.Fatalf("write repo config: %v", err)
	}

	cfg, err := Load(repoRoot)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.Defaults.Repo != "repo/repo" || cfg.Defaults.PR != 222 {
		t.Fatalf("expected repo config to override: %+v", cfg.Defaults)
	}
	if !cfg.Defaults.IncludeDiff {
		t.Fatalf("expected global default to remain when repo config omits it")
	}
}
