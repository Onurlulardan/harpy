package publication

import (
	"testing"
	"time"
)

func TestPublicReleaseRequiresDifferentFreshStepUpApprover(t *testing.T) {
	id, _ := NewID(v7(1))
	requester := v7(2)
	digest := [32]byte{1}
	release, err := NewRelease(id, v7(3), AudiencePublic, CacheRevisionBound, digest, requester)
	if err != nil {
		t.Fatal(err)
	}
	release, err = release.Submit(1)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	if _, err := release.Approve(2, requester, now, now); err != ErrConflict {
		t.Fatalf("same-principal approval error = %v", err)
	}
	release, err = release.Approve(2, v7(4), now.Add(-9*time.Minute), now)
	if err != nil {
		t.Fatal(err)
	}
	release, err = release.Activate(3)
	if err != nil || release.State() != StateActive {
		t.Fatalf("activate result = %#v, %v", release, err)
	}
	if _, err := release.Revoke(3); err != ErrConflict {
		t.Fatalf("stale revoke error = %v", err)
	}
}

func v7(seed byte) [16]byte {
	value := [16]byte{0, 0, 0, 0, 0, 0, 0x70, seed, 0x80}
	value[15] = seed
	return value
}
