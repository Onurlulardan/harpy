package main

import (
	"testing"
	"time"
)

func TestReuseEvidenceClockKeepsValidityFloor(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 18, 18, 12, 0, 0, time.UTC)
	got := reuseEvidenceClock(now)
	if got.Sub(now) != reusePrepareValidityFloor ||
		reusePrepareValidityFloor != 2*time.Minute {
		t.Fatalf("reuse evidence clock = %s, floor=%s", got, reusePrepareValidityFloor)
	}
}
