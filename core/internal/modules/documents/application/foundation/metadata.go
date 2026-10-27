package foundation

import (
	"encoding/json"
	"sort"
	"strings"
	"unicode/utf8"
)

func NormalizeItemMetadata(description string, tags []string) (string, []string, []byte, bool) {
	description = strings.TrimSpace(description)
	if !utf8.ValidString(description) || len([]byte(description)) > 4096 || len(tags) > 32 {
		return "", nil, nil, false
	}
	result := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, value := range tags {
		value = strings.TrimSpace(value)
		if value == "" || !utf8.ValidString(value) || len([]byte(value)) > 128 {
			return "", nil, nil, false
		}
		if _, exists := seen[value]; exists {
			return "", nil, nil, false
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	encoded, err := json.Marshal(result)
	if err != nil || len(encoded) > 2048 {
		return "", nil, nil, false
	}
	return description, result, encoded, true
}
