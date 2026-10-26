package physicalid_test

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/physicalid"
	"github.com/google/uuid"
)

func TestNameMatchesAcceptedUUIDMapping(t *testing.T) {
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	name, err := physicalid.Name("rel_", id)
	if err != nil || name != "rel_aaisem2ekvthpcezvk54zxpo74" {
		t.Fatalf("unexpected physical relation name: %q, %v", name, err)
	}
}

func TestNameRejectsUnknownPrefixOrInvalidUUID(t *testing.T) {
	valid := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	if _, err := physicalid.Name("raw_", valid); err == nil {
		t.Fatal("unknown physical identifier prefix accepted")
	}
	if _, err := physicalid.Name("rel_", uuid.Nil); err == nil {
		t.Fatal("zero physical identifier UUID accepted")
	}
}

func FuzzNameIsCanonicalAndBounded(f *testing.F) {
	f.Add([]byte("0123456789abcdef"), "rel_")
	f.Fuzz(func(t *testing.T, raw []byte, prefix string) {
		if len(raw) != 16 {
			return
		}
		id, err := uuid.FromBytes(raw)
		if err != nil {
			return
		}
		name, err := physicalid.Name(prefix, id)
		if err != nil {
			return
		}
		if !strings.HasPrefix(name, prefix) || len(name) != len(prefix)+26 ||
			name != strings.ToLower(name) {
			t.Fatal("non-canonical physical identifier produced")
		}
	})
}
