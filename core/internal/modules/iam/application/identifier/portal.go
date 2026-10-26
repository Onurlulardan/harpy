package identifier

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const maximumPortalScalars = 100

func NormalizePortalV1(raw string) (Value, error) {
	if !utf8.ValidString(raw) {
		return Value{}, ErrInvalid
	}
	display := strings.TrimFunc(raw, unicode.IsSpace)
	if display == "" {
		return Value{}, ErrInvalid
	}
	normalized := unicodeDefaultCasefoldV1(display)
	if normalized == "" || utf8.RuneCountInString(normalized) > maximumPortalScalars {
		return Value{}, ErrInvalid
	}
	for _, character := range normalized {
		if unicode.IsControl(character) {
			return Value{}, ErrInvalid
		}
	}
	return newValue(PortalV1, display, normalized), nil
}
