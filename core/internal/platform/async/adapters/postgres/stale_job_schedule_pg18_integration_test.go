//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

func assertFutureStaleAdoptionSchedule(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
) {
	t.Helper()
	tenantID := addStaleJobTenant(t, ctx, fixture)
	jobID, sourceRunAt := insertFutureLifecycleJob(t, ctx, fixture, tenantID)
	setPublisherDomainGeneration(t, ctx, fixture.publisherPGFixture, tenantID, 2)
	if parked, err := fixture.store.ReconcileStaleJobs(ctx, 50); err != nil || parked != 1 {
		t.Fatalf("park future stale job = %d, %v", parked, err)
	}
	authority, err := fixture.store.ResolveStaleJobAuthority(ctx, jobID)
	if err != nil {
		t.Fatal(err)
	}
	command := staleDispositionCommand(
		t, jobID, jobsv1.StaleJobAdopt, "adopt future database work",
	)
	result, err := fixture.store.DisposeStaleJob(ctx, authority, command)
	if err != nil || !result.Created() {
		t.Fatalf("future adoption = %#v, %v", result, err)
	}
	resultID, ok := result.ResultJobID()
	if !ok {
		t.Fatal("future adoption omitted replacement")
	}
	var resultRunAt time.Time
	if err := fixture.owner.QueryRowContext(ctx,
		"SELECT run_at FROM platform.jobs WHERE id=$1", resultID,
	).Scan(&resultRunAt); err != nil {
		t.Fatal(err)
	}
	if !resultRunAt.Equal(sourceRunAt) {
		t.Fatalf("future adoption schedule = %s, want %s", resultRunAt, sourceRunAt)
	}
	claims, err := fixture.store.ClaimJobs(
		ctx, authority, uuid.New(), 1,
		lifecycleCatalog(t, jobsv1.EffectDatabaseTransactional, 8),
	)
	if err != nil || len(claims) != 0 {
		t.Fatalf("future replacement claimed early = %#v, %v", claims, err)
	}
	wait := time.Until(sourceRunAt) + 100*time.Millisecond
	if wait > 0 {
		timer := time.NewTimer(wait)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-timer.C:
		}
	}
	claims, err = fixture.store.ClaimJobs(
		ctx, authority, uuid.New(), 1,
		lifecycleCatalog(t, jobsv1.EffectDatabaseTransactional, 8),
	)
	if err != nil || len(claims) != 1 || claims[0].JobID() != resultID {
		t.Fatalf("future replacement claim at schedule = %#v, %v", claims, err)
	}
}

func insertFutureLifecycleJob(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	tenantID uuid.UUID,
) (uuid.UUID, time.Time) {
	t.Helper()
	jobID := uuid.New()
	var runAt time.Time
	err := fixture.runtime.QueryRow(ctx, `
INSERT INTO platform.jobs (
    id,tenant_id,scope_kind,scope_id,authority_group_id,
    authority_group_generation,authority_domain,cutover_generation,writer_origin,
    kind,payload_revision,payload_ref,payload_sha256,dedupe_key,effect_class,
    priority,run_at,state,attempts,max_attempts,lease_generation
) VALUES (
    $1,$2,'tenant',$2,'operations',1,'workflow',1,'v2',
    'workflow.execute',1,$3,repeat('a',64),$4,'database_transactional',10,
    pg_catalog.clock_timestamp()+interval '3 seconds','queued',0,8,0
) RETURNING run_at`, jobID, tenantID, "payload/"+jobID.String(),
		"dedupe:"+jobID.String()).Scan(&runAt)
	if err != nil {
		t.Fatalf("insert future lifecycle job: %v", err)
	}
	return jobID, runAt
}
