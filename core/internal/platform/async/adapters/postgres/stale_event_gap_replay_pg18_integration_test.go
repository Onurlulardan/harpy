//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestStaleEventGapDispositionReplayOnPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			fixture := newEventGapPGFixture(t, ctx, origin)
			assertStaleEventGapDispositionReplay(t, ctx, &fixture)
		})
	}
}

func assertStaleEventGapDispositionReplay(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
) {
	t.Helper()
	for sequence := int64(1); sequence <= 3; sequence++ {
		fixture.seedDelivered(t, ctx, sequence)
	}
	fixture.deliver(t, ctx, 3)
	_, jobID := fixture.gapIDs(t, ctx, 3)
	setPublisherDomainGeneration(
		t, ctx, fixture.publisherPGFixture, fixture.tenantID, 2,
	)
	if parked, err := fixture.store.ReconcileStaleJobs(ctx, 50); err != nil || parked != 1 {
		t.Fatalf("park stale event-gap job = %d, %v", parked, err)
	}
	command := staleDispositionCommand(
		t, jobID, jobsv1.StaleJobAdopt, "adopt stale event gap",
	)
	authority, err := fixture.store.ResolveStaleJobAuthority(ctx, jobID)
	if err != nil {
		t.Fatal(err)
	}
	created, err := fixture.store.DisposeStaleJob(ctx, authority, command)
	if err != nil || !created.Created() {
		t.Fatalf("adopt stale event gap = %#v, %v", created, err)
	}
	rootID, ok := created.ResultRootID()
	if !ok || rootID.Version() != 8 {
		t.Fatalf("stale event-gap result root = %s/%t", rootID, ok)
	}
	replayed, err := fixture.store.DisposeStaleJob(ctx, authority, command)
	if err != nil {
		t.Fatalf("replay stale event gap: %v", err)
	}
	assertSameStaleDispositionResult(t, created, replayed, "event gap")
}
