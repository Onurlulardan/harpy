package domain

import (
	"errors"
	"testing"
)

func TestParseTenantIDReadsNetworkOrderBytes(t *testing.T) {
	resolved, err := ParseTenantID("00112233-4455-6677-8899-aabbccddeeff")
	if err != nil {
		t.Fatal(err)
	}
	want := TenantID{
		0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77,
		0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff,
	}
	if resolved != want {
		t.Fatalf("tenant ID bytes = %x, want %x", resolved, want)
	}
}

func TestParseTenantIDRejectsNonCanonicalValues(t *testing.T) {
	tests := []string{
		"", "00000000-0000-0000-0000-000000000000",
		"00112233445566778899aabbccddeeff",
		"00112233-4455-6677-8899-AABBCCDDEEFF",
		"00112233-4455-6677-8899-aabbccddeefg",
	}
	for _, input := range tests {
		if _, err := ParseTenantID(input); !errors.Is(err, ErrInvalidTenantID) {
			t.Fatalf("ParseTenantID(%q) error = %v", input, err)
		}
	}
}
