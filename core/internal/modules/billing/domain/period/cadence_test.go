package period

import (
	"errors"
	"testing"
	"time"
)

func TestBuildManualRejectsUnknownCadence(t *testing.T) {
	start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	anchor, err := AnchorFrom(start)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := BuildManual(start, anchor, Cadence("weekly")); !errors.Is(err, ErrCadence) {
		t.Fatalf("unknown cadence error = %v", err)
	}
}
