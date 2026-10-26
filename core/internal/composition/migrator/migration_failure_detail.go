package migrator

import (
	"strings"
)

func migrationFailureDetail(err error) string {
	if err == nil {
		return ""
	}
	text := err.Error()
	if index := strings.Index(text, "ERROR:"); index >= 0 {
		text = text[index:]
	}
	if index := strings.Index(text, " (SQLSTATE "); index >= 0 {
		if end := strings.Index(text[index:], ")"); end >= 0 {
			text = text[:index+end+1]
		}
	}
	if len(text) > 400 {
		text = text[:400]
	}
	lower := strings.ToLower(text)
	if strings.Contains(lower, "postgres://") || strings.Contains(text, "SELECT ") ||
		strings.Contains(text, `C:\\`) || strings.Contains(text, "/secret") {
		return "migration execution failed"
	}
	return text
}
