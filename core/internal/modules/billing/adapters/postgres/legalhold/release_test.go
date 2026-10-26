package legalhold

import (
	"context"
	"testing"
	"time"

	legalholdapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/legalhold"
	"github.com/google/uuid"
)

func TestReleaseReturnsPersistedLifecycleEvidence(t *testing.T) {
	tenantID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	actorID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	authorityID := uuid.MustParse("018f1000-0000-7000-8000-000000000110")
	holdID := uuid.MustParse("018f1000-0000-7000-8000-000000000111")
	placeID := uuid.MustParse("018f1000-0000-7000-8000-000000000112")
	releaseID := uuid.MustParse("018f1000-0000-7000-8000-000000000113")
	releasedAt := adapterNow.Add(time.Hour)
	row := activeRow(holdID, tenantID, actorID, placeID)
	row[2], row[8], row[9] = "released", nullableUUID{value: actorID, Valid: true},
		nullableUUID{value: releaseID, Valid: true}
	row[10] = nullableText{value: "court released preservation order", Valid: true}
	row[11], row[12] = nullableTime{value: releasedAt, Valid: true}, int64(2)
	queries := 0
	storage, err := NewStore(tenantID, DatabaseAccess{Allowed: func(context.Context) bool {
		return true
	}, QueryRow: func(
		_ context.Context, statement string, _ ...any,
	) Row {
		queries++
		if statement == resolveAuthoritySQL {
			return rowStub{values: []any{authorityID}}
		}
		return rowStub{values: row}
	}})
	if err != nil {
		t.Fatal(err)
	}
	result, err := storage.Release(context.Background(), legalholdapp.ReleaseMutation{
		OperationID: releaseID, TenantID: tenantID, HoldID: holdID, ActorID: actorID,
		Key: "release-LEGAL-2026-41", RequestSHA: make([]byte, 32),
		Reason: "court released preservation order", OccurredAt: releasedAt,
	})
	if err != nil || queries != 2 || result.ReleaseOperationID != releaseID {
		t.Fatalf("release result=%+v queries=%d err=%v", result, queries, err)
	}
}
