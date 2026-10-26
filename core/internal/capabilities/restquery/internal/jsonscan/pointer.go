package jsonscan

import "strings"

func appendPointer(pointer, token string) string {
	escaped := strings.ReplaceAll(token, "~", "~0")
	escaped = strings.ReplaceAll(escaped, "/", "~1")
	return pointer + "/" + escaped
}
