//go:build integration

package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func assertPublisherOrderingAndLifecycle(
	t *testing.T,
	ctx context.Context,
	fixture publisherPGFixture,
) {
	t.Helper()
	tenantA := uuid.MustParse(asyncTenantA)
	tenantB := uuid.MustParse(asyncTenantB)
	aggregateA, aggregateB := uuid.New(), uuid.New()
	firstID, secondID, tenantBID := uuid.New(), uuid.New(), uuid.New()
	seedPublisherEvent(t, ctx, fixture, publisherEventSeed{
		eventID: firstID, tenantID: tenantA, aggregateID: aggregateA, sequence: 1,
	})
	seedPublisherEvent(t, ctx, fixture, publisherEventSeed{
		eventID: secondID, tenantID: tenantA, aggregateID: aggregateA, sequence: 2,
	})
	seedPublisherEvent(t, ctx, fixture, publisherEventSeed{
		eventID: tenantBID, tenantID: tenantB, aggregateID: aggregateB, sequence: 1,
	})
	authorityA := publisherAuthorityFor(t, ctx, fixture, tenantA)
	publisherID := uuid.New()
	claims, err := fixture.store.ClaimOutbox(
		ctx, authorityA, publisherID, 50, fixture.catalog,
	)
	if err != nil || len(claims) != 1 || claims[0].eventID != firstID {
		t.Fatalf("first stream claim = %#v, %v", claims, err)
	}
	claim := claims[0]
	lease, err := fixture.store.HeartbeatPublish(ctx, claim)
	if err != nil || !lease.Until().Equal(lease.HeartbeatAt().Add(publishLeaseDuration)) ||
		!lease.HeartbeatAt().After(claim.Lease().HeartbeatAt()) {
		t.Fatalf("heartbeat lease = %#v, %v", lease, err)
	}
	if _, err := fixture.store.ReleasePublish(
		ctx, claim, 0, "BROKER_NACK", strings.Repeat("a", 64),
	); err != nil {
		t.Fatalf("release first claim: %v", err)
	}
	assertReleasedPublisherClaim(t, ctx, fixture, claim.eventID)
	if _, err := fixture.store.ConfirmPublish(
		ctx, claim, strings.Repeat("b", 64),
	); !errors.Is(err, ErrPublishLeaseLost) {
		t.Fatalf("released claim confirmation error = %v", err)
	}
	claims, err = fixture.store.ClaimOutbox(
		ctx, authorityA, uuid.New(), 50, fixture.catalog,
	)
	if err != nil || len(claims) != 1 || claims[0].eventID != firstID ||
		claims[0].LeaseGeneration() != 2 {
		t.Fatalf("reclaim = %#v, %v", claims, err)
	}
	confirmPublisherClaim(t, ctx, fixture, claims[0], "c")
	claims, err = fixture.store.ClaimOutbox(
		ctx, authorityA, uuid.New(), 50, fixture.catalog,
	)
	if err != nil || len(claims) != 1 || claims[0].eventID != secondID {
		t.Fatalf("second stream claim = %#v, %v", claims, err)
	}
	confirmPublisherClaim(t, ctx, fixture, claims[0], "d")
	authorityB := publisherAuthorityFor(t, ctx, fixture, tenantB)
	claims, err = fixture.store.ClaimOutbox(
		ctx, authorityB, uuid.New(), 50, fixture.catalog,
	)
	if err != nil || len(claims) != 1 || claims[0].eventID != tenantBID ||
		claims[0].Envelope().TenantID == nil ||
		*claims[0].Envelope().TenantID != tenantB.String() {
		t.Fatalf("tenant B claim = %#v, %v", claims, err)
	}
	confirmPublisherClaim(t, ctx, fixture, claims[0], "e")
}

func confirmPublisherClaim(
	t *testing.T,
	ctx context.Context,
	fixture publisherPGFixture,
	claim PublishClaim,
	digestCharacter string,
) {
	t.Helper()
	deliveredAt, err := fixture.store.ConfirmPublish(
		ctx, claim, strings.Repeat(digestCharacter, 64),
	)
	if err != nil || deliveredAt.IsZero() {
		t.Fatalf("confirm %s: %v, %v", claim.eventID, deliveredAt, err)
	}
	var state, digest string
	var leaseCleared, timestampsMatch bool
	if err := fixture.owner.QueryRowContext(ctx, `
SELECT delivery_state, confirm_sha256,
       lease_owner_id IS NULL AND lease_heartbeat_at IS NULL AND lease_until IS NULL,
       confirmed_at = delivered_at AND delivered_at = $2
FROM platform.outbox WHERE event_id = $1`, claim.eventID, deliveredAt).Scan(
		&state, &digest, &leaseCleared, &timestampsMatch,
	); err != nil {
		t.Fatal(err)
	}
	if state != "delivered" || digest != strings.Repeat(digestCharacter, 64) ||
		!leaseCleared || !timestampsMatch {
		t.Fatalf("confirmed row = %s, %s, %v, %v", state, digest, leaseCleared, timestampsMatch)
	}
}

func assertReleasedPublisherClaim(
	t *testing.T,
	ctx context.Context,
	fixture publisherPGFixture,
	eventID uuid.UUID,
) {
	t.Helper()
	var state, code, digest string
	var leaseCleared bool
	if err := fixture.owner.QueryRowContext(ctx, `
SELECT delivery_state, last_error_code, last_error_sha256,
       lease_owner_id IS NULL AND lease_heartbeat_at IS NULL AND lease_until IS NULL
FROM platform.outbox WHERE event_id = $1`, eventID).Scan(
		&state, &code, &digest, &leaseCleared,
	); err != nil {
		t.Fatal(err)
	}
	if state != "pending" || code != "BROKER_NACK" || digest != strings.Repeat("a", 64) ||
		!leaseCleared {
		t.Fatalf("released row = %s, %s, %s, %v", state, code, digest, leaseCleared)
	}
}
