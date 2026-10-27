package domain

import (
	"crypto/sha256"
	"errors"
	"testing"
	"time"
)

func TestTrashRestoreAndRetentionPurge(t *testing.T) {
	item := validItem(t)
	now := time.Date(2026, 7, 18, 8, 0, 0, 0, time.UTC)
	trashed, err := item.Trash(now)
	if err != nil || trashed.RetentionEndsAt != now.Add(TrashRetention) {
		t.Fatalf("trash = %+v, %v", trashed, err)
	}
	if _, err = trashed.BeginRetentionPurge(trashed.RetentionEndsAt, true); !errors.Is(err, ErrLegalHold) {
		t.Fatalf("legal hold error = %v", err)
	}
	restored, err := trashed.Restore(now.Add(time.Hour))
	if err != nil || restored.State != ItemActive || !restored.RetentionEndsAt.IsZero() {
		t.Fatalf("restore = %+v, %v", restored, err)
	}
}

func TestManualPurgeCoolingCancelAndDeleteEvidence(t *testing.T) {
	item := validItem(t)
	now := time.Date(2026, 7, 18, 8, 0, 0, 0, time.UTC)
	cooling, err := item.RequestPurge(now, now.Add(-5*time.Minute), "duplicate upload")
	if err != nil || cooling.PurgeStartsAt != now.Add(PurgeCooling) {
		t.Fatalf("request = %+v, %v", cooling, err)
	}
	if _, err = cooling.BeginManualPurge(now.Add(PurgeCooling-time.Second), false); !errors.Is(err, ErrItemTransition) {
		t.Fatalf("early purge error = %v", err)
	}
	cancelled, err := cooling.CancelPurge(now.Add(time.Hour))
	if err != nil || cancelled.State != ItemActive {
		t.Fatalf("cancel = %+v, %v", cancelled, err)
	}
	purging, _ := cooling.BeginManualPurge(cooling.PurgeStartsAt, false)
	evidence := sha256.Sum256([]byte("all exact object versions deleted"))
	if _, err = purging.FinishPurge(cooling.PurgeStartsAt, 1, evidence); !errors.Is(err, ErrItemTransition) {
		t.Fatalf("pending object error = %v", err)
	}
	purged, err := purging.FinishPurge(cooling.PurgeStartsAt, 0, evidence)
	if err != nil || purged.State != ItemPurged ||
		purged.TombstoneUntil != cooling.PurgeStartsAt.Add(TombstoneKeep) {
		t.Fatalf("purged = %+v, %v", purged, err)
	}
}

func TestManualPurgeRequiresFreshStepUp(t *testing.T) {
	item := validItem(t)
	now := time.Date(2026, 7, 18, 8, 0, 0, 0, time.UTC)
	if _, err := item.RequestPurge(now, now.Add(-StepUpMaximumAge-time.Second), "reason"); !errors.Is(err, ErrItemTransition) {
		t.Fatalf("stale step-up error = %v", err)
	}
}

func validItem(t *testing.T) Item {
	t.Helper()
	item, err := NewItem(
		mustID(t, "018f0000-0000-7000-8000-000000000010"),
		mustID(t, "11111111-1111-4111-8111-111111111111"),
		mustID(t, "018f0000-0000-7000-8000-000000000011"),
		mustID(t, "018f0000-0000-7000-8000-000000000012"),
	)
	if err != nil {
		t.Fatal(err)
	}
	return item
}
