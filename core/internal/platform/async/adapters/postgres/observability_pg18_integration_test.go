//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAsyncObservabilityAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			fixture := newPublisherPGFixture(t, ctx, origin)
			assertAsyncObservabilitySemantics(t, ctx, fixture)
		})
	}
}

func assertAsyncObservabilitySemantics(
	t *testing.T, ctx context.Context, fixture publisherPGFixture,
) {
	t.Helper()
	tenantA, tenantB := uuid.MustParse(asyncTenantA), uuid.MustParse(asyncTenantB)
	oldestOutboxID := uuid.New()
	seedPublisherEvent(t, ctx, fixture, publisherEventSeed{
		eventID: oldestOutboxID, tenantID: tenantA,
		aggregateID: uuid.New(), sequence: 1,
	})
	source := &observabilitySource{database: fixture.runtime}
	initial, err := source.Snapshot(ctx)
	if err != nil {
		t.Fatalf("observe initial async backlog: %v", err)
	}
	if alarm := initial.UndeliveredOutboxAlarm(); alarm != 0 {
		t.Fatalf("fresh outbox alarm = %d, want 0", alarm)
	}
	seedObservabilityJob(t, ctx, fixture, tenantA, time.Now().UTC().Add(-90*time.Second))
	parkedOutboxID := uuid.New()
	seedPublisherEvent(t, ctx, fixture, publisherEventSeed{
		eventID: parkedOutboxID, tenantID: tenantB,
		aggregateID: uuid.New(), sequence: 1,
	})
	seedObservabilityJob(t, ctx, fixture, tenantB, time.Now().UTC())
	advancePublisherAuthority(t, ctx, fixture, tenantB, 2)
	authority := publisherAuthorityFor(t, ctx, fixture, tenantB)
	if parked, err := fixture.store.ParkStaleOutbox(ctx, authority, 1); err != nil || parked != 1 {
		t.Fatalf("park observability outbox = %d, %v", parked, err)
	}
	worker := &workerRepository{
		database: fixture.runtime, cleanup: 5 * time.Second,
		identity: &workerIdentity{marker: 1},
	}
	if parked, err := worker.ReconcileStaleJobs(ctx, 50); err != nil || parked != 1 {
		t.Fatalf("park observability job = %d, %v", parked, err)
	}
	ageOutboxForObservation(t, ctx, fixture, oldestOutboxID, 301)
	snapshot, err := source.Snapshot(ctx)
	if err != nil {
		t.Fatalf("observe async backlog: %v", err)
	}
	if snapshot.OldestUndeliveredOutboxAgeSeconds() < 301 ||
		snapshot.UndeliveredOutboxAlarm() != 1 ||
		snapshot.OldestDueJobLagSeconds() < 90 ||
		snapshot.ParkedJobs() != 1 || snapshot.ParkedOutbox() != 1 {
		t.Fatalf("async PG observation snapshot = %#v", snapshot)
	}
	assertObservabilityRuntimeRole(t, ctx, fixture)
}
