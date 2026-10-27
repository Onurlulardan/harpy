//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func assertPublisherStaleParking(
	t *testing.T, ctx context.Context, fixture publisherPGFixture,
) {
	t.Helper()
	tenantID := uuid.MustParse(asyncTenantA)
	publishingID, pendingID := uuid.New(), uuid.New()
	seedPublisherEvent(t, ctx, fixture, publisherEventSeed{
		eventID: publishingID, tenantID: tenantID,
		aggregateID: uuid.New(), sequence: 1,
	})
	authority := publisherAuthorityFor(t, ctx, fixture, tenantID)
	claims, err := fixture.store.ClaimOutbox(
		ctx, authority, uuid.New(), 1, fixture.catalog,
	)
	if err != nil || len(claims) != 1 || claims[0].eventID != publishingID {
		t.Fatalf("claim stale publishing event = %#v, %v", claims, err)
	}
	expirePublisherLease(t, ctx, fixture, publishingID)
	seedPublisherEvent(t, ctx, fixture, publisherEventSeed{
		eventID: pendingID, tenantID: tenantID,
		aggregateID: uuid.New(), sequence: 1,
	})
	advancePublisherAuthority(t, ctx, fixture, tenantID, 2)
	current := publisherAuthorityFor(t, ctx, fixture, tenantID)
	claims, err = fixture.store.ClaimOutbox(
		ctx, current, uuid.New(), 50, fixture.catalog,
	)
	if err != nil || len(claims) != 0 {
		t.Fatalf("current generation claimed stale events = %#v, %v", claims, err)
	}
	parked, err := fixture.store.ParkStaleOutbox(ctx, current, 1)
	if err != nil || parked != 1 {
		t.Fatalf("first stale park = %d, %v", parked, err)
	}
	restarted := fixture
	restarted.runtime = openPublisherRuntimePool(t, ctx, fixture.owner)
	restarted.store = &workerRepository{
		database: restarted.runtime, cleanup: 5 * time.Second,
		identity: &workerIdentity{marker: 1},
	}
	restartedAuthority := publisherAuthorityFor(t, ctx, restarted, tenantID)
	parked, err = restarted.store.ParkStaleOutbox(ctx, restartedAuthority, 50)
	if err != nil || parked != 1 {
		t.Fatalf("post-restart stale park = %d, %v", parked, err)
	}
	assertStaleOutboxEvidence(t, ctx, fixture, publishingID, pendingID)
}
