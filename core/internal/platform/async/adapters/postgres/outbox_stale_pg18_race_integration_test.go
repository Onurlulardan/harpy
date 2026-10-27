//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func assertPublisherStaleParkingRace(
	t *testing.T, ctx context.Context, fixture publisherPGFixture,
) {
	t.Helper()
	tenantID := uuid.MustParse(asyncTenantB)
	eventID := uuid.New()
	seedPublisherEvent(t, ctx, fixture, publisherEventSeed{
		eventID: eventID, tenantID: tenantID,
		aggregateID: uuid.New(), sequence: 1,
	})
	advancePublisherAuthority(t, ctx, fixture, tenantID, 2)
	stores := []*workerRepository{
		{database: fixture.runtime, cleanup: 5 * time.Second,
			identity: &workerIdentity{marker: 1}},
		{database: fixture.runtime, cleanup: 5 * time.Second,
			identity: &workerIdentity{marker: 1}},
	}
	authorities := make([]OutboxAuthority, len(stores))
	for index, store := range stores {
		view := fixture
		view.store = store
		authorities[index] = publisherAuthorityFor(t, ctx, view, tenantID)
	}
	type outcome struct {
		parked int64
		err    error
	}
	start := make(chan struct{})
	results := make(chan outcome, len(stores))
	for index, store := range stores {
		go func(store *workerRepository, authority OutboxAuthority) {
			<-start
			parked, err := store.ParkStaleOutbox(ctx, authority, 1)
			results <- outcome{parked: parked, err: err}
		}(store, authorities[index])
	}
	close(start)
	var total int64
	for range stores {
		result := <-results
		if result.err != nil {
			t.Fatalf("racing stale park: %v", result.err)
		}
		total += result.parked
	}
	if total != 1 {
		t.Fatalf("racing stale park total = %d, want 1", total)
	}
	var parkedRows, evidenceRows int64
	if err := fixture.owner.QueryRowContext(ctx, `
SELECT count(*) FILTER (WHERE parked_at IS NOT NULL),
       (SELECT count(*) FROM platform.dead_letters WHERE source_id=$1)
FROM platform.outbox WHERE event_id=$1`, eventID).Scan(
		&parkedRows, &evidenceRows,
	); err != nil {
		t.Fatal(err)
	}
	if parkedRows != 1 || evidenceRows != 1 {
		t.Fatalf("race result parked/evidence = %d/%d", parkedRows, evidenceRows)
	}
}
