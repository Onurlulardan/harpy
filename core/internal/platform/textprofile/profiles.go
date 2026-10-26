package textprofile

import (
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// ExactNFC applies the exact, case-sensitive normalization profile.
func ExactNFC(value string) string {
	return norm.NFC.String(value)
}

// UnicodeDefaultCasefoldV1 applies the versioned locale-independent fold.
func UnicodeDefaultCasefoldV1(value string) string {
	return ExactNFC(cases.Fold().String(ExactNFC(value)))
}

// AccentInsensitiveV1 applies casefold V1 and removes every Unicode mark class.
func AccentInsensitiveV1(value string) string {
	decomposed := norm.NFD.String(UnicodeDefaultCasefoldV1(value))
	var result strings.Builder
	result.Grow(len(decomposed))
	for _, character := range decomposed {
		if unicode.Is(unicode.Mn, character) ||
			unicode.Is(unicode.Mc, character) ||
			unicode.Is(unicode.Me, character) {
			continue
		}
		result.WriteRune(character)
	}
	return ExactNFC(result.String())
}
