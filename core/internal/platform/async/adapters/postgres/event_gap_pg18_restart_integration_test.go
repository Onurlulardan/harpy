//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEventGapReconcileCommitCrashRestartsSafelyOnPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
			defer cancel()
			fixture := newEventGapPGFixture(t, ctx, origin)
			assertEventGapRestart(t, ctx, &fixture)
		})
	}
}

func assertEventGapRestart(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
) {
	t.Helper()
	fixture.seedDelivered(t, ctx, 1)
	fixture.seedDelivered(t, ctx, 2)
	fixture.deliver(t, ctx, 2)
	_, jobID := fixture.gapIDs(t, ctx, 2)
	fixture.deliver(t, ctx, 1)
	claim := fixture.claimGap(t, ctx, jobID)
	installGapReconcileCommitCrash(t, ctx, fixture, jobID)
	if _, err := fixture.store.ReconcileEventGap(ctx, claim); !pipelineCommitCrash(err) {
		t.Fatalf("injected reconcile commit crash = %v", err)
	}
	assertGapReconcileRolledBack(t, ctx, fixture, jobID)
	removeGapReconcileCommitCrash(t, ctx, fixture)
	expireGapJobLease(t, ctx, fixture, jobID)

	restarted := &workerRepository{
		database: fixture.runtime, cleanup: 5 * time.Second,
		identity: &workerIdentity{marker: 1},
	}
	authority := eventGapAuthorityForStore(t, ctx, restarted, fixture.tenantID)
	recovered, err := restarted.RecoverExpiredJobs(ctx, authority, 1)
	if err != nil || recovered.Requeued() != 1 {
		t.Fatalf("restart recovery = %#v, %v", recovered, err)
	}
	claims, err := restarted.ClaimJobs(
		ctx, authority, uuid.New(), 1,
		consumerJobCatalogForPipeline(fixture.consumers),
	)
	if err != nil || len(claims) != 1 || claims[0].JobID() != jobID ||
		claims[0].Attempt() != 2 {
		t.Fatalf("restart claim = %#v, %v", claims, err)
	}
	result, err := restarted.ReconcileEventGap(ctx, claims[0])
	if err != nil || !result.Requeued() || result.ActionID() != jobID {
		t.Fatalf("restart reconcile = %#v, %v", result, err)
	}
	republishGap(t, ctx, fixture, 2)
	assertGapEffects(t, ctx, fixture, 2)
}
