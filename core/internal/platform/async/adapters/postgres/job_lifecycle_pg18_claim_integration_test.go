//go:build integration

package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

func assertJobClaimHeartbeatAndCancellation(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
) {
	t.Helper()
	jobA := insertQueuedLifecycleJob(
		t, ctx, fixture, asyncTenantA, jobsv1.EffectDatabaseTransactional, 8,
	)
	jobB := insertQueuedLifecycleJob(
		t, ctx, fixture, asyncTenantB, jobsv1.EffectDatabaseTransactional, 8,
	)
	workerID := uuid.New()
	claims, err := fixture.store.ClaimJobs(
		ctx, fixture.authority, workerID, 10,
		lifecycleCatalog(t, jobsv1.EffectDatabaseTransactional, 8),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(claims) != 1 || claims[0].JobID() != jobA || claims[0].JobID() == jobB {
		t.Fatalf("tenant-isolated claims = %#v", claims)
	}
	lease, err := fixture.store.HeartbeatJob(ctx, claims[0])
	if err != nil || lease.Until().Sub(lease.HeartbeatAt()) != jobLeaseDuration {
		t.Fatalf("heartbeat lease = %#v, %v", lease, err)
	}
	cancellation, err := NewJobCancellation(
		"USER_CANCELLED", strings.Repeat("c", 64),
	)
	if err != nil {
		t.Fatal(err)
	}
	control, err := fixture.store.ResolveJobControl(ctx, fixture.authority, jobA)
	if err != nil {
		t.Fatal(err)
	}
	requested, err := fixture.store.RequestJobCancellation(ctx, control, cancellation)
	if err != nil || requested.State() != jobsv1.StateRunning || requested.Terminal() {
		t.Fatalf("running cancellation = %#v, %v", requested, err)
	}
	lease, err = fixture.store.HeartbeatJob(ctx, claims[0])
	if err != nil || !lease.CancelRequested() {
		t.Fatalf("cooperative cancellation heartbeat = %#v, %v", lease, err)
	}
	if _, err := fixture.store.CancelJob(ctx, claims[0], cancellation); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.store.HeartbeatJob(
		ctx, claims[0],
	); !errors.Is(err, ErrJobLeaseLost) {
		t.Fatalf("stale heartbeat error = %v", err)
	}
	state, open := lifecycleState(t, ctx, fixture, jobA)
	if state != string(jobsv1.StateCancelled) || open != 0 {
		t.Fatalf("cancelled state = %s, open attempts = %d", state, open)
	}
}
