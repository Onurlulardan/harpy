package textprofile

import (
	"testing"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

func FuzzProfilesRemainCanonicalAndIdempotent(f *testing.F) {
	for _, seed := range []string{
		"Cafe\u0301", "Stra\u00dfe", "I\u0130\u0131i", "a\u0903a\u20dd", "\x00",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value string) {
		if !utf8.ValidString(value) {
			return
		}
		for _, profile := range []func(string) string{
			ExactNFC, UnicodeDefaultCasefoldV1, AccentInsensitiveV1,
		} {
			canonical := profile(value)
			if !norm.NFC.IsNormalString(canonical) || profile(canonical) != canonical {
				t.Fatalf("profile output is not canonical and idempotent: %q", canonical)
			}
		}
		decomposed := norm.NFD.String(AccentInsensitiveV1(value))
		for _, character := range decomposed {
			if unicode.Is(unicode.Mn, character) || unicode.Is(unicode.Mc, character) ||
				unicode.Is(unicode.Me, character) {
				t.Fatalf("accent profile retained mark %U", character)
			}
		}
	})
}
