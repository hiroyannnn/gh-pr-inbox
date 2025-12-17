package config

import (
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
	if !cfg.Defaults.IncludeIssueComment || !cfg.Defaults.NoUpdateCheck {
		t.Fatalf("unexpected defaults: %+v", cfg.Defaults)
	}
}
