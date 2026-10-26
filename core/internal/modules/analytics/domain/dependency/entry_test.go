package dependency

import (
	"strings"
	"testing"
)

func TestEntryAllowsNoDataRevisionCapability(t *testing.T) {
	var resourceID, versionID [16]byte
	resourceID[0], versionID[0] = 1, 2
	var digest [32]byte
	digest[0] = 3
	if _, err := NewEntry(
		"forms", "form", resourceID, versionID,
		1, 1, 1, "", digest,
	); err != nil {
		t.Fatalf("empty revision capability rejected: %v", err)
	}
	if _, err := NewEntry(
		"forms", "form", resourceID, versionID,
		1, 1, 1, strings.Repeat("r", 129), digest,
	); err == nil {
		t.Fatal("oversized revision capability accepted")
	}
}
