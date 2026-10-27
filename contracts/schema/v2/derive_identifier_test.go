package schemav2

import (
	"strings"
	"testing"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func TestDeriveFieldIdentifierPinsVariants(t *testing.T) {
	wants := map[FieldVariant]string{
		VariantValue:  "fld_aghup5d3dn6mbaaaaaaaaaaham_g1",
		VariantFolded: "fld_aghup5d3dn6mbaaaaaaaaaaham_g1_folded",
		VariantAccent: "fld_aghup5d3dn6mbaaaaaaaaaaham_g1_accent",
	}
	for variant, want := range wants {
		actual, err := DeriveFieldIdentifier(testFieldID, 1, variant)
		if err != nil || actual != want {
			t.Fatalf("variant %s = %q, %v", variant, actual, err)
		}
	}
	v1Value, err := schemav1.DeriveFieldIdentifier(testFieldID, 1)
	if err != nil || v1Value != wants[VariantValue] {
		t.Fatalf("V1 value identifier drifted: %q, %v", v1Value, err)
	}
}

func TestDeriveFieldIdentifierRejectsOpenVariant(t *testing.T) {
	for _, variant := range []FieldVariant{"", "value;DROP TABLE", "raw"} {
		if _, err := DeriveFieldIdentifier(testFieldID, 1, variant); err == nil {
			t.Fatalf("variant accepted: %q", variant)
		}
	}
}

func FuzzDeriveFieldIdentifierNeverLeaksInput(f *testing.F) {
	f.Add(testFieldID, int64(1), string(VariantFolded))
	f.Add("slug;drop table", int64(1), "raw")
	f.Fuzz(func(t *testing.T, id string, generation int64, variant string) {
		name, err := DeriveFieldIdentifier(id, generation, FieldVariant(variant))
		if err != nil {
			return
		}
		if len(name) > 63 || strings.ContainsAny(name, ` ;"'`) ||
			!strings.HasPrefix(name, "fld_") {
			t.Fatalf("unsafe identifier %q", name)
		}
	})
}
