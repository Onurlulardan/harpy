package providerapp

import (
	"unicode"
	"unicode/utf8"
)

const maximumExternalIdentifierBytes = 512

func validExternalIdentifier(value string) bool {
	if value == "" || len(value) > maximumExternalIdentifierBytes || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}
