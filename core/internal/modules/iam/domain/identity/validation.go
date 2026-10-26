package identity

import "strings"

func ValidDisplayName(value string) bool {
	return len(value) >= 1 && len(value) <= 255 && strings.TrimSpace(value) == value
}

func ValidSlug(value string) bool {
	if len(value) < 1 || len(value) > 100 || value != strings.ToLower(value) {
		return false
	}
	for index, character := range value {
		valid := character >= 'a' && character <= 'z' ||
			character >= '0' && character <= '9' || character == '-'
		if !valid || character == '-' && (index == 0 || index == len(value)-1) {
			return false
		}
	}
	return true
}

func ValidProfileInput(value []byte) bool {
	return len(value) >= 2 && len(value) <= MaximumProfileBytes
}
