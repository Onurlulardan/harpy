package main

import (
	"testing"
	"time"
)

func generatedSetForTest(t *testing.T) (*artifactSet, time.Time) {
	t.Helper()
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	set, err := generateArtifactSet(now)
	if err != nil {
		t.Fatalf("generate artifact set: %v", err)
	}
	return set, now
}
