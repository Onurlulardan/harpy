package identifier

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizePortalV1CanonicalizesUnicode(t *testing.T) {
	t.Parallel()
	value, err := NormalizePortalV1("\u2003 Stra\u00dfe-42 \u00a0")
	if err != nil {
		t.Fatal(err)
	}
	if !value.Valid() || value.Kind() != PortalV1 ||
		value.Display() != "Stra\u00dfe-42" || value.Normalized() != "strasse-42" {
		t.Fatalf("portal value = %#v", value)
	}
}

func TestNormalizePortalV1RejectsInvalidValues(t *testing.T) {
	t.Parallel()
	tests := []string{
		"", " \t\r\n", "code\x00value", string([]byte{0xff}),
		strings.Repeat("a", maximumPortalScalars+1),
	}
	for _, input := range tests {
		if value, err := NormalizePortalV1(input); !errors.Is(err, ErrInvalid) ||
			value.Valid() {
			t.Fatalf("invalid portal accepted: %q %#v %v", input, value, err)
		}
	}
}

func TestNormalizePortalV1PreservesInternalSpace(t *testing.T) {
	t.Parallel()
	value, err := NormalizePortalV1(" A  B ")
	if err != nil || value.Normalized() != "a  b" {
		t.Fatalf("portal value = %#v, %v", value, err)
	}
}
