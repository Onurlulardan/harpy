package legalhold

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestReleasePersistsTypedLegalHoldEvidence(t *testing.T) {
	tenantID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	actorID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	holdID := uuid.MustParse("018f1000-0000-7000-8000-000000000101")
	placeID := uuid.MustParse("018f1000-0000-7000-8000-000000000102")
	releaseID := uuid.MustParse("018f1000-0000-7000-8000-000000000103")
	snapshot := activeSnapshot(holdID, placeID, tenantID, actorID)
	snapshot.State, snapshot.Revision = "released", 2
	snapshot.ReleasedBy, snapshot.ReleaseOperationID = actorID, releaseID
	snapshot.ReleaseReason = "court released preservation order"
	snapshot.ReleasedAt = legalHoldNow.Add(time.Hour)
	store := &storeStub{released: snapshot}
	commands, err := NewCommands(store, idSequence(releaseID), func() time.Time {
		return legalHoldNow.Add(time.Hour)
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := commands.Release(context.Background(), ReleaseRequest{
		TenantID: tenantID, HoldID: holdID, ActorID: actorID,
		IdempotencyKey: "release-LEGAL-2026-41",
		Reason:         "court released preservation order",
	})
	if err != nil || result.ReleaseOperationID != releaseID ||
		store.release.OperationID != releaseID || len(store.release.RequestSHA) != 32 {
		t.Fatalf("release result=%+v mutation=%+v err=%v", result, store.release, err)
	}
}

func TestReleaseRejectsInvalidHoldIdentity(t *testing.T) {
	commands, err := NewCommands(&storeStub{}, uuid.NewV7, func() time.Time {
		return legalHoldNow
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = commands.Release(context.Background(), ReleaseRequest{})
	if !errors.Is(err, ErrRequest) {
		t.Fatalf("invalid release error=%v", err)
	}
}
