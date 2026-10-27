package compiler

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

func validAggregateOutputKey(value string, measure bool) bool {
	if value == "" || value == "_included" || !utf8.ValidString(value) ||
		strings.ContainsRune(value, '\x00') || !norm.NFC.IsNormalString(value) {
		return false
	}
	if !measure {
		return true
	}
	if len(value) > 63 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for index := 1; index < len(value); index++ {
		character := value[index]
		if (character < 'a' || character > 'z') &&
			(character < '0' || character > '9') && character != '_' {
			return false
		}
	}
	return true
}
