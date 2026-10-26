package attestation

import (
	"testing"
	"time"
)

func TestDecodeStrictTreatsTimeAsRFC3339Scalar(t *testing.T) {
	t.Parallel()
	type document struct {
		ObservedAt time.Time `json:"observed_at"`
	}
	var value document
	if err := DecodeStrict([]byte(`{"observed_at":"2026-07-23T10:00:00Z"}`), &value); err != nil {
		t.Fatal(err)
	}
	if !value.ObservedAt.Equal(time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("observed_at = %s", value.ObservedAt)
	}
	if DecodeStrict([]byte(`{"observed_at":{}}`), &value) == nil {
		t.Fatal("object time was accepted")
	}
}
