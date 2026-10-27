package published

import "strings"

func parameterPointer(root, name string) string {
	return root + "/" + escapePointerToken(name)
}

func escapePointerToken(token string) string {
	escaped := strings.ReplaceAll(token, "~", "~0")
	return strings.ReplaceAll(escaped, "/", "~1")
}
