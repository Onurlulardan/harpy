package domain

import "testing"

func TestParseIDRoundTripsCanonicalUUIDv7(t *testing.T) {
	const value = "018f9f3e-7c16-7a21-8b42-116a9d3c9870"
	id, err := ParseID(value)
	if err != nil || !id.UUIDv7() || id.String() != value {
		t.Fatalf("ParseID = %q, %v", id.String(), err)
	}
	if _, err := ParseID("018f9f3e-7c16-7a21-8b42-116a9d3c987z"); err == nil {
		t.Fatal("ParseID accepted malformed input")
	}
}
