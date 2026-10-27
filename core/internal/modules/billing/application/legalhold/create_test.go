package legalhold

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCreatePersistsTypedLegalHoldEvidence(t *testing.T) {
	tenantID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	actorID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	holdID := uuid.MustParse("018f1000-0000-7000-8000-000000000101")
	operationID := uuid.MustParse("018f1000-0000-7000-8000-000000000102")
	store := &storeStub{placed: activeSnapshot(
		holdID, operationID, tenantID, actorID,
	)}
	commands, err := NewCommands(store, idSequence(holdID, operationID), func() time.Time {
		return legalHoldNow
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := commands.Create(context.Background(), PlaceRequest{
		TenantID: tenantID, ActorID: actorID, IdempotencyKey: "hold-LEGAL-2026-41",
		Reason: "litigation preservation", Ticket: "LEGAL-2026-41",
	})
	if err != nil || result.HoldID != holdID || store.place.HoldID != holdID ||
		len(store.place.RequestSHA) != 32 || store.place.OccurredAt != legalHoldNow {
		t.Fatalf("place result=%+v mutation=%+v err=%v", result, store.place, err)
	}
}

func TestCreateRejectsUnboundedOrInvalidInput(t *testing.T) {
	commands, err := NewCommands(&storeStub{}, uuid.NewV7, func() time.Time {
		return legalHoldNow
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = commands.Create(context.Background(), PlaceRequest{})
	if !errors.Is(err, ErrRequest) {
		t.Fatalf("invalid place error=%v", err)
	}
}
