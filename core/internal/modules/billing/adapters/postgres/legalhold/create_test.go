package legalhold

import (
	"context"
	"testing"

	legalholdapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/legalhold"
	"github.com/google/uuid"
)

func TestCreateUsesCurrentTenantAuthority(t *testing.T) {
	tenantID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	actorID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	authorityID := uuid.MustParse("018f1000-0000-7000-8000-000000000110")
	holdID := uuid.MustParse("018f1000-0000-7000-8000-000000000111")
	operationID := uuid.MustParse("018f1000-0000-7000-8000-000000000112")
	queries := 0
	bound, err := NewStore(tenantID, DatabaseAccess{Allowed: func(context.Context) bool {
		return true
	}, QueryRow: func(
		_ context.Context, statement string, _ ...any,
	) Row {
		queries++
		if statement == resolveAuthoritySQL {
			return rowStub{values: []any{authorityID}}
		}
		return rowStub{values: activeRow(holdID, tenantID, actorID, operationID)}
	}})
	if err != nil {
		t.Fatal(err)
	}
	storage := bound.(*store)
	result, err := storage.Create(context.Background(), legalholdapp.PlaceMutation{
		HoldID: holdID, OperationID: operationID, TenantID: tenantID,
		ActorID: actorID, Key: "hold-LEGAL-2026-41", RequestSHA: make([]byte, 32),
		Reason: "litigation preservation", Ticket: "LEGAL-2026-41",
		OccurredAt: adapterNow,
	})
	if err != nil || queries != 2 || result.HoldID != holdID {
		t.Fatalf("place result=%+v queries=%d err=%v", result, queries, err)
	}
}
