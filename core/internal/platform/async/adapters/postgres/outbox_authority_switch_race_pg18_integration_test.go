//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestOutboxAuthorityPredicateSwitchRaceOnPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			fixture := newPublisherPGFixture(t, ctx, origin)
			assertOutboxAuthoritySwitchRace(t, ctx, fixture)
		})
	}
}

func assertOutboxAuthoritySwitchRace(
	t *testing.T,
	ctx context.Context,
	fixture publisherPGFixture,
) {
	t.Helper()
	tenantA := uuid.MustParse(asyncTenantA)
	tenantB := uuid.MustParse(asyncTenantB)
	eventA := seedAuthorityRaceOutbox(t, ctx, fixture, tenantA)
	eventB := seedAuthorityRaceOutbox(t, ctx, fixture, tenantB)
	assertPredicateBlocksAuthorityUpdate(t, ctx, fixture.owner, eventA, `
UPDATE platform.cutover_groups SET generation=2
WHERE scope_kind='tenant' AND scope_id=$1 AND group_id='operations'`, tenantA)
	assertPredicateBlocksAuthorityUpdate(t, ctx, fixture.owner, eventB, `
UPDATE platform.domain_cutovers SET generation=2
WHERE scope_kind='tenant' AND scope_id=$1 AND domain='workflow'`, tenantB)
}

func seedAuthorityRaceOutbox(
	t *testing.T,
	ctx context.Context,
	fixture publisherPGFixture,
	tenantID uuid.UUID,
) uuid.UUID {
	t.Helper()
	eventID := uuid.New()
	seedPublisherEvent(t, ctx, fixture, publisherEventSeed{
		eventID: eventID, tenantID: tenantID,
		aggregateID: uuid.New(), sequence: 1,
	})
	authority := publisherAuthorityFor(t, ctx, fixture, tenantID)
	claim := claimPipelineEvent(t, ctx, fixture, authority, eventID, 1)
	if _, err := fixture.store.ConfirmPublish(
		ctx, claim, eventGapDigest("race-confirm:"+eventID.String()),
	); err != nil {
		t.Fatal(err)
	}
	return eventID
}

func assertPredicateBlocksAuthorityUpdate(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	eventID uuid.UUID,
	update string,
	scopeID uuid.UUID,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	var current bool
	if err := transaction.QueryRowContext(ctx, `
SELECT platform.platform_async_outbox_authority_current_v1(event)
FROM platform.outbox AS event WHERE event.event_id=$1`, eventID).Scan(&current); err != nil {
		t.Fatal(err)
	}
	if !current {
		t.Fatal("current outbox predicate returned stale")
	}
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		close(started)
		_, updateErr := database.ExecContext(ctx, update, scopeID)
		done <- updateErr
	}()
	<-started
	select {
	case updateErr := <-done:
		t.Fatalf("authority switch escaped predicate lock: %v", updateErr)
	case <-time.After(200 * time.Millisecond):
	}
	var asserted any
	if err := transaction.QueryRowContext(ctx, `
SELECT platform.platform_async_assert_current_authority_v1(
    event.scope_kind,event.scope_id,event.authority_group_id,
    event.authority_group_generation,event.authority_domain,
    event.cutover_generation)
FROM platform.outbox AS event WHERE event.event_id=$1`, eventID).Scan(&asserted); err != nil {
		t.Fatalf("same-XID authority assertion: %v", err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	select {
	case updateErr := <-done:
		if updateErr != nil {
			t.Fatalf("authority switch after predicate commit: %v", updateErr)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("authority switch remained blocked after predicate commit")
	}
}
