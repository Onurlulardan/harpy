package httpquery

import (
	"strings"

	"golang.org/x/text/unicode/norm"
)

func splitGETName(name string) ([]string, bool) {
	open := strings.IndexByte(name, '[')
	if open < 0 {
		return []string{name}, name != "" && !strings.ContainsRune(name, ']')
	}
	if open == 0 || strings.ContainsRune(name[:open], ']') {
		return nil, false
	}
	segments := []string{name[:open]}
	position := open
	for position < len(name) {
		if name[position] != '[' {
			return nil, false
		}
		closeOffset := strings.IndexByte(name[position+1:], ']')
		if closeOffset < 0 {
			return nil, false
		}
		closeAt := position + 1 + closeOffset
		segment := name[position+1 : closeAt]
		if strings.ContainsRune(segment, '[') {
			return nil, false
		}
		segments = append(segments, segment)
		position = closeAt + 1
	}
	return segments, true
}

func getNamePointer(name string) string {
	segments, valid := splitGETName(name)
	if !valid {
		return ""
	}
	pointer := ""
	for _, segment := range segments {
		pointer = appendJSONPointer(pointer, segment)
	}
	return pointer
}

func getSegmentsAreNFC(segments []string) bool {
	for _, segment := range segments {
		if !norm.NFC.IsNormalString(segment) {
			return false
		}
	}
	return true
}
