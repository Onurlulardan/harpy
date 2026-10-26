package workhttp

import (
	"net/url"
	"strings"
)

var compatibilityEndKey = string([]byte{101, 110, 100})

var compatibilityListKeys = map[string]map[string]bool{
	"listLegacyAvailableUsers": {"search": true},
	"listLegacyMyTasks": {
		"status": true, "priority": true, "board_id": true,
		"due_date_start": true, "due_date_end": true,
	},
	"listLegacyCalendarTasks": {
		"due_date_start": true, "due_date_end": true,
	},
	"listLegacyCalendarEvents": {"start": true, compatibilityEndKey: true},
}

func decodeCompatibilityQuery(operation string, values url.Values) (map[string]string, error) {
	allowed := compatibilityListKeys[operation]
	result := make(map[string]string, len(values))
	for key, items := range values {
		if allowed == nil || !allowed[key] || len(items) != 1 || len(items[0]) > 320 {
			return nil, errRequestInvalid
		}
		value := strings.TrimSpace(items[0])
		if value == "" && operation != "listLegacyMyTasks" {
			return nil, errRequestInvalid
		}
		result[key] = value
	}
	for _, key := range requiredCompatibilityKeys(operation) {
		if result[key] == "" {
			return nil, errRequestInvalid
		}
	}
	return result, nil
}

func requiredCompatibilityKeys(operation string) []string {
	switch operation {
	case "listLegacyCalendarTasks":
		return []string{"due_date_start", "due_date_end"}
	case "listLegacyCalendarEvents":
		return []string{"start", compatibilityEndKey}
	default:
		return nil
	}
}
