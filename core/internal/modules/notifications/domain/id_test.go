package domain

import (
	"bytes"
	"testing"
	"time"
)

func TestNewV7IsCanonicalAndRoundTrips(t *testing.T) {
	wantTime := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	value, err := NewV7(wantTime, bytes.NewReader(bytes.Repeat([]byte{0x2a}, 10)))
	if err != nil || !value.IsV7() {
		t.Fatalf("NewV7() = %v, %v", value, err)
	}
	decoded, err := ParseID(value.String())
	if err != nil || decoded != value {
		t.Fatalf("ParseID() = %v, %v", decoded, err)
	}
}

func TestParseIDRejectsZeroAndMalformedValues(t *testing.T) {
	values := []string{"", "00000000-0000-0000-0000-000000000000", "not-an-id"}
	for _, value := range values {
		if _, err := ParseID(value); err == nil {
			t.Fatalf("ParseID(%q) accepted", value)
		}
	}
}
