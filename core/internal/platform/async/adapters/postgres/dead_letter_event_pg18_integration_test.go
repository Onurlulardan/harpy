//go:build integration

package postgres

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	"github.com/google/uuid"
)

func TestDeadLetterEventOperatorAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			fixture := newDeadLetterPGFixture(t, ctx, origin)
			assertDeadLetterEventOperator(t, ctx, fixture)
		})
	}
}

func assertDeadLetterEventOperator(
	t *testing.T,
	ctx context.Context,
	fixture deadLetterPGFixture,
) {
	t.Helper()
	tenantA := uuid.MustParse(asyncTenantA)
	tenantB := uuid.MustParse(asyncTenantB)
	rootA := deadLetterEventRoot(t, tenantA, uuid.New(), uuid.New(), fixture.consumers)
	rootB := deadLetterEventRoot(t, tenantB, uuid.New(), uuid.New(), fixture.consumers)
	first, err := fixture.repository.IngestEvent(ctx, rootA)
	if err != nil || !first.Created {
		t.Fatalf("ingest tenant A dead-letter = %#v, %v", first, err)
	}
	if _, err := fixture.repository.IngestEvent(ctx, rootB); err != nil {
		t.Fatal(err)
	}
	duplicate, err := fixture.repository.IngestEvent(ctx, rootA)
	if err != nil || duplicate.Created || duplicate.RootID != first.RootID {
		t.Fatalf("duplicate ingest = %#v, %v", duplicate, err)
	}
	assertDeadLetterReadIsolation(t, ctx, fixture, rootA, rootB)
	assertDeadLetterEventRetry(t, ctx, fixture, rootA)
}

func assertDeadLetterReadIsolation(
	t *testing.T,
	ctx context.Context,
	fixture deadLetterPGFixture,
	rootA deadletter.EventRoot,
	rootB deadletter.EventRoot,
) {
	t.Helper()
	tenantA, tenantB := *rootA.TenantID(), *rootB.TenantID()
	tenantOperator, _ := deadletter.NewOperator(
		deadLetterPermissionFixture{allowed: map[uuid.UUID]struct{}{tenantA: {}}},
		fixture.repository, deadLetterFixedID{value: uuid.New()},
	)
	page, err := tenantOperator.List(ctx, deadletter.ListRequest{
		ActorID: uuid.New(), TenantIDs: []uuid.UUID{tenantA}, Limit: 10,
	})
	if err != nil || len(page.Entries) != 1 || page.Entries[0].TenantID != tenantA {
		t.Fatalf("tenant dead-letter page = %#v, %v", page, err)
	}
	_, err = tenantOperator.Retry(ctx, deadletter.RetryRequest{
		ActorID: uuid.New(), TenantID: tenantB, RootID: rootB.RootID(),
		Reason: "idor probe", Ticket: "SEC-42", CausationID: uuid.New(),
	})
	if !errors.Is(err, deadletter.ErrUnauthorized) {
		t.Fatalf("cross-tenant retry error = %v", err)
	}
	platformOperator, _ := deadletter.NewOperator(
		deadLetterPermissionFixture{
			allowed: map[uuid.UUID]struct{}{tenantA: {}, tenantB: {}}, platform: true,
		}, fixture.repository, deadLetterFixedID{value: uuid.New()},
	)
	page, err = platformOperator.List(ctx, deadletter.ListRequest{
		ActorID: uuid.New(), TenantIDs: []uuid.UUID{tenantA, tenantB},
		Platform: true, Limit: 10,
	})
	if err != nil || len(page.Entries) != 2 {
		t.Fatalf("explicit platform dead-letter page = %#v, %v", page, err)
	}
}

func assertDeadLetterEventRetry(
	t *testing.T,
	ctx context.Context,
	fixture deadLetterPGFixture,
	root deadletter.EventRoot,
) {
	t.Helper()
	tenantID := *root.TenantID()
	seedDeadLetterEffectReceipt(t, ctx, fixture, tenantID)
	before := readDeadLetterPayload(t, ctx, fixture, root.RootID())
	actionID := uuid.New()
	operator, _ := deadletter.NewOperator(
		deadLetterPermissionFixture{allowed: map[uuid.UUID]struct{}{tenantID: {}}},
		fixture.repository, deadLetterFixedID{value: actionID},
	)
	result, err := operator.Retry(ctx, deadletter.RetryRequest{
		ActorID: uuid.New(), TenantID: tenantID, RootID: root.RootID(),
		Reason: "consumer upgraded", Ticket: "OPS-42", CausationID: uuid.New(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.SourceKind != deadletter.SourceEvent || result.ResultID != result.SourceID ||
		result.DeliveryJobID == nil || *result.DeliveryJobID != actionID {
		t.Fatalf("event retry receipt = %#v", result)
	}
	after := readDeadLetterPayload(t, ctx, fixture, root.RootID())
	if !bytes.Equal(before, after) {
		t.Fatal("manual event retry mutated dead-letter payload")
	}
	assertDeadLetterEventRetryRows(t, ctx, fixture, root, actionID)
}
