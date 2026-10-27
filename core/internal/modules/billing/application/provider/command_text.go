package providerapp

import (
	"strings"
	"unicode/utf8"
)

func validCommandText(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && utf8.ValidString(value) &&
		strings.TrimSpace(value) == value
}
