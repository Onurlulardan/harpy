package identifier

import (
	"testing"
	"unicode"

	"golang.org/x/net/idna"
)

func TestIdentifierV1UnicodeTablesStayPinned(t *testing.T) {
	t.Parallel()
	if unicode.Version != UnicodeDataVersionV1 ||
		idna.UnicodeVersion != UnicodeDataVersionV1 {
		t.Fatalf("Unicode tables = go:%s idna:%s, want %s",
			unicode.Version, idna.UnicodeVersion, UnicodeDataVersionV1)
	}
}
