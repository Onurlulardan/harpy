package v1query

import "strings"

func appendPointer(root, token string) string {
	escaped := strings.ReplaceAll(token, "~", "~0")
	escaped = strings.ReplaceAll(escaped, "/", "~1")
	return root + "/" + escaped
}

func queryKeyPointer(key string) string {
	segments, valid := splitV1Key(key)
	if !valid || len(segments) == 0 {
		return appendPointer("", key)
	}
	result := ""
	for _, segment := range segments {
		result = appendPointer(result, segment)
	}
	return result
}
