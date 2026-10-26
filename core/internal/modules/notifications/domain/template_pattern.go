package domain

import "strings"

func templatePatternsValid(subject, body string, allowed map[string]struct{}) bool {
	used := make(map[string]struct{}, len(allowed))
	for _, pattern := range []string{subject, body} {
		for {
			start := strings.Index(pattern, "{{")
			if start < 0 {
				if strings.Contains(pattern, "}}") {
					return false
				}
				break
			}
			if strings.Contains(pattern[:start], "}}") {
				return false
			}
			end := strings.Index(pattern[start+2:], "}}")
			if end < 0 {
				return false
			}
			end += start + 2
			name := pattern[start+2 : end]
			if _, exists := allowed[name]; !exists || !validVariable(name) {
				return false
			}
			used[name] = struct{}{}
			pattern = pattern[end+2:]
		}
	}
	return len(used) == len(allowed)
}
