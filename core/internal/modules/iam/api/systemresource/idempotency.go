package systemresource

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const maximumIdempotencyKeyBytes = 128

func validIdempotencyKey(value string) bool {
	if value == "" || len(value) > maximumIdempotencyKeyBytes ||
		!utf8.ValidString(value) || strings.TrimSpace(value) != value {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}
