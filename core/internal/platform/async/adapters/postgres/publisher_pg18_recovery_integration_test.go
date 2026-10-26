//go:build integration

package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func assertPublisherLeaseRecovery(
	t *testing.T,
	ctx context.Context,
	fixture publisherPGFixture,
) {
	t.Helper()
	tenantID := uuid.MustParse(asyncTenantA)
	eventID := uuid.New()
	seedPublisherEvent(t, ctx, fixture, publisherEventSeed{
		eventID: eventID, tenantID: tenantID,
		aggregateID: uuid.New(), sequence: 1,
	})
	authority := publisherAuthorityFor(t, ctx, fixture, tenantID)
	claims, err := fixture.store.ClaimOutbox(
		ctx, authority, uuid.New(), 1, fixture.catalog,
	)
	if err != nil || len(claims) != 1 {
		t.Fatalf("claim recovery event = %#v, %v", claims, err)
	}
	stale := claims[0]
	expirePublisherLease(t, ctx, fixture, eventID)
	if _, err := fixture.store.HeartbeatPublish(
		ctx, stale,
	); !errors.Is(err, ErrPublishLeaseLost) {
		t.Fatalf("expired heartbeat error = %v", err)
	}
	if _, err := fixture.store.ConfirmPublish(
		ctx, stale, strings.Repeat("1", 64),
	); !errors.Is(err, ErrPublishLeaseLost) {
		t.Fatalf("expired confirm error = %v", err)
	}
	recovered, err := fixture.store.RecoverExpiredOutbox(ctx, authority, 50)
	if err != nil || recovered != 1 {
		t.Fatalf("recovered = %d, %v", recovered, err)
	}
	assertRecoveredPublisherRow(t, ctx, fixture, eventID)
	claims, err = fixture.store.ClaimOutbox(
		ctx, authority, uuid.New(), 1, fixture.catalog,
	)
	if err != nil || len(claims) != 1 || claims[0].LeaseGeneration() != 2 {
		t.Fatalf("post-recovery claim = %#v, %v", claims, err)
	}
	if _, err := fixture.store.ConfirmPublish(
		ctx, stale, strings.Repeat("2", 64),
	); !errors.Is(err, ErrPublishLeaseLost) {
		t.Fatalf("stale generation confirm error = %v", err)
	}
	confirmPublisherClaim(t, ctx, fixture, claims[0], "3")
}

func expirePublisherLease(
	t *testing.T,
	ctx context.Context,
	fixture publisherPGFixture,
	eventID uuid.UUID,
) {
	t.Helper()
	transaction, err := fixture.owner.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err := transaction.ExecContext(
		ctx, "ALTER TABLE platform.outbox DISABLE TRIGGER outbox_transition_guard",
	); err != nil {
		t.Fatalf("disable publisher transition guard: %v", err)
	}
	if _, err := transaction.ExecContext(ctx, `UPDATE platform.outbox SET
        lease_heartbeat_at = pg_catalog.clock_timestamp() - interval '61 seconds',
        lease_until = pg_catalog.clock_timestamp() - interval '1 second',
        updated_at = pg_catalog.clock_timestamp()
      WHERE event_id = $1`, eventID); err != nil {
		t.Fatalf("expire publisher lease: %v", err)
	}
	if _, err := transaction.ExecContext(
		ctx, "ALTER TABLE platform.outbox ENABLE ALWAYS TRIGGER outbox_transition_guard",
	); err != nil {
		t.Fatalf("enable publisher transition guard: %v", err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func assertRecoveredPublisherRow(
	t *testing.T,
	ctx context.Context,
	fixture publisherPGFixture,
	eventID uuid.UUID,
) {
	t.Helper()
	var state, code, digest string
	if err := fixture.owner.QueryRowContext(ctx, `
SELECT delivery_state, last_error_code, last_error_sha256
FROM platform.outbox WHERE event_id = $1`, eventID).Scan(&state, &code, &digest); err != nil {
		t.Fatal(err)
	}
	if state != "pending" || code != "OUTBOX_LEASE_EXPIRED" ||
		!publisherDigestPattern.MatchString(digest) {
		t.Fatalf("recovered state = %s, %s, %s", state, code, digest)
	}
}
