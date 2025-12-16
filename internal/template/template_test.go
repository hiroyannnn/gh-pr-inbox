package template

import "testing"

func TestApplySubstitutesVariables(t *testing.T) {
	prompt := "Hello {{NAME}}, PR {{NUMBER}}"
	vars := map[string]string{"NAME": "world", "NUMBER": "42"}

	out := Apply(prompt, vars)
	if out != "Hello world, PR 42" {
		t.Fatalf("unexpected substitution result: %s", out)
	}
}
