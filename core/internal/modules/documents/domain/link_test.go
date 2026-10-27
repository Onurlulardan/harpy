package domain

import (
	"errors"
	"testing"
	"time"
)

func TestLinkPinsSourceAndEpochAndCannotReactivate(t *testing.T) {
	now := time.Date(2026, 7, 18, 10, 0, 0, 0, time.UTC)
	link, err := NewLink(mustID(t, "018f0000-0000-7000-8000-000000000301"),
		mustID(t, "11111111-1111-4111-8111-111111111111"),
		mustID(t, "018f0000-0000-7000-8000-000000000302"),
		mustID(t, "33333333-3333-4333-8333-333333333333"),
		mustID(t, "44444444-4444-4444-8444-444444444444"),
		"FORMS", "record", "READ", 7, now)
	if err != nil || link.ReferenceEpoch != 7 {
		t.Fatalf("link = %+v, %v", link, err)
	}
	released, err := link.Release(now.Add(time.Second))
	if err != nil || released.State != LinkReleased {
		t.Fatalf("released = %+v, %v", released, err)
	}
	if _, err = released.Release(now.Add(2 * time.Second)); !errors.Is(err, ErrLinkInvalid) {
		t.Fatalf("second release error = %v", err)
	}
}
