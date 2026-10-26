package domain

import "testing"

func TestIDFormatsCanonicalUUIDAndExposesVersion(t *testing.T) {
	id := ID{0x01, 0x90, 0, 0, 0, 0, 0x70, 0, 0x80, 0, 0, 0, 0, 0, 0, 2}
	if id.String() != "01900000-0000-7000-8000-000000000002" || id.Version() != 7 {
		t.Fatalf("unexpected support ID %s version %d", id.String(), id.Version())
	}
}
