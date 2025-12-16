package template

import "strings"

// Apply substitutes stable variables in the prompt template.
func Apply(prompt string, vars map[string]string) string {
	replacer := strings.NewReplacer(toPairs(vars)...)
	return replacer.Replace(prompt)
}

func toPairs(vars map[string]string) []string {
	pairs := make([]string, 0, len(vars)*2)
	for k, v := range vars {
		pairs = append(pairs, "{{"+k+"}}", v)
	}
	return pairs
}
