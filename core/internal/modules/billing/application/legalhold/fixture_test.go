package legalhold

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type storeStub struct {
	placed   Snapshot
	released Snapshot
	place    PlaceMutation
	release  ReleaseMutation
	err      error
}

func (stub *storeStub) Place(
	_ context.Context,
	mutation PlaceMutation,
) (Snapshot, error) {
	stub.place = mutation
	return stub.placed, stub.err
}

func (stub *storeStub) Release(
	_ context.Context,
	mutation ReleaseMutation,
) (Snapshot, error) {
	stub.release = mutation
	return stub.released, stub.err
}

var legalHoldNow = time.Date(2026, 7, 17, 10, 0, 0, 0, time.UTC)

func idSequence(values ...uuid.UUID) IDSource {
	index := 0
	return func() (uuid.UUID, error) {
		value := values[index]
		index++
		return value, nil
	}
}

func activeSnapshot(
	holdID, operationID, tenantID, actorID uuid.UUID,
) Snapshot {
	return Snapshot{
		HoldID: holdID, TenantID: tenantID, State: "active", PlacedBy: actorID,
		PlaceOperationID: operationID, Reason: "litigation preservation",
		Ticket: "LEGAL-2026-41", PlacedAt: legalHoldNow, Revision: 1,
	}
}
