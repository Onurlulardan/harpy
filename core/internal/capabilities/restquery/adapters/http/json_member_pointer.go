package httpquery

import "strings"

func appendJSONPointer(pointer, token string) string {
	return pointer + "/" + escapeJSONPointerToken(token)
}

func escapeJSONPointerToken(token string) string {
	escaped := strings.ReplaceAll(token, "~", "~0")
	return strings.ReplaceAll(escaped, "/", "~1")
}
