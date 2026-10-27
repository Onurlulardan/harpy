package schemav1

import "testing"

func TestDeriveIdentifierPinsNetworkOrderBase32(t *testing.T) {
	const id = "018f47f4-7b1b-7cc0-8000-000000000702"
	tests := []struct {
		kind IdentifierKind
		want string
	}{
		{IdentifierForm, "frm_aghup5d3dn6mbaaaaaaaaaahai"},
		{IdentifierRelation, "rel_aghup5d3dn6mbaaaaaaaaaahai"},
		{IdentifierIndex, "idx_aghup5d3dn6mbaaaaaaaaaahai"},
		{IdentifierConstraint, "cst_aghup5d3dn6mbaaaaaaaaaahai"},
	}
	for _, test := range tests {
		actual, err := DeriveIdentifier(test.kind, id)
		if err != nil || actual != test.want {
			t.Fatalf("DeriveIdentifier(%s) = %q, %v", test.kind, actual, err)
		}
	}
	field, err := DeriveFieldIdentifier(id, 36)
	if err != nil || field != "fld_aghup5d3dn6mbaaaaaaaaaahai_g10" {
		t.Fatalf("field identifier = %q, %v", field, err)
	}
}

func TestDeriveIdentifierRejectsOpenInputs(t *testing.T) {
	for _, value := range []string{
		"", "018F47F4-7B1B-7CC0-8000-000000000702",
		"00000000-0000-0000-0000-000000000000", "slug",
	} {
		if _, err := DeriveIdentifier(IdentifierForm, value); err == nil {
			t.Fatalf("invalid logical ID accepted: %q", value)
		}
	}
	if _, err := DeriveIdentifier("raw_prefix", testIdentifier()); err == nil {
		t.Fatal("unknown identifier kind accepted")
	}
	for _, generation := range []int64{0, -1, MaximumFieldGeneration + 1} {
		if _, err := DeriveFieldIdentifier(testIdentifier(), generation); err == nil {
			t.Fatalf("generation accepted: %d", generation)
		}
	}
}

func testIdentifier() string { return "018f47f4-7b1b-7cc0-8000-000000000702" }
