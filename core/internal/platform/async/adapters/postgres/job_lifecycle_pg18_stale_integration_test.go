//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

func assertStaleJobParkAndRestart(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
) {
	t.Helper()
	tenantID := addStaleJobTenant(t, ctx, fixture)
	authority, err := newJobAuthority(
		fixture.store.identity, "tenant", tenantID,
		"operations", 1, "workflow", 1,
	)
	if err != nil {
		t.Fatal(err)
	}
	fixture.authority = authority
	runningID := insertQueuedLifecycleJob(
		t, ctx, fixture, tenantID.String(), jobsv1.EffectDatabaseTransactional, 8,
	)
	claims, err := fixture.store.ClaimJobs(
		ctx, fixture.authority, uuid.New(), 1,
		lifecycleCatalog(t, jobsv1.EffectDatabaseTransactional, 8),
	)
	if err != nil || len(claims) != 1 || claims[0].JobID() != runningID {
		t.Fatalf("claim stale running job = %#v, %v", claims, err)
	}
	legacyClaim, legacyPayload, legacyDigest := claimLegacyStaleJob(
		t, ctx, fixture, tenantID,
	)
	queuedID := insertQueuedLifecycleJob(
		t, ctx, fixture, tenantID.String(), jobsv1.EffectDatabaseTransactional, 8,
	)
	effectID := insertQueuedLifecycleJob(
		t, ctx, fixture, tenantID.String(), jobsv1.EffectExternal, 8,
	)
	effectClaim := claimOneExternalJob(t, ctx, fixture, effectID)
	_ = markExternalEffect(t, ctx, fixture, effectClaim)

	setPublisherDomainGeneration(
		t, ctx, fixture.publisherPGFixture, tenantID, 2,
	)
	if _, err := fixture.store.HeartbeatJob(
		ctx, claims[0],
	); !errors.Is(err, ErrJobAuthorityStale) {
		t.Fatalf("stale heartbeat error = %v", err)
	}
	if parked, err := fixture.store.ParkStaleJob(ctx, claims[0]); err != nil || !parked {
		t.Fatalf("park claimed stale job = %t, %v", parked, err)
	}
	assertLegacyStaleJobPark(
		t, ctx, fixture, legacyClaim, legacyPayload, legacyDigest,
	)
	if parked, err := fixture.store.ParkStaleJob(
		ctx, effectClaim,
	); err != nil || parked {
		t.Fatalf("effect-started stale job park = %t, %v", parked, err)
	}
	if parked, err := fixture.store.ReconcileStaleJobs(ctx, 50); err != nil || parked != 2 {
		t.Fatalf("reconcile stale jobs = %d, %v", parked, err)
	}
	if parked, err := fixture.store.ReconcileStaleJobs(ctx, 50); err != nil || parked != 0 {
		t.Fatalf("restart stale reconciliation = %d, %v", parked, err)
	}
	assertStaleJobRows(t, ctx, fixture, runningID, queuedID, effectID)
}

func addStaleJobTenant(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
) uuid.UUID {
	t.Helper()
	tenantID := uuid.New()
	_, err := fixture.owner.ExecContext(ctx, `
WITH tenant_row AS (
    INSERT INTO platform.tenants (id) VALUES ($1) RETURNING id
), group_row AS (
    INSERT INTO platform.cutover_groups (
    scope_kind,scope_id,group_id,state,writer,generation
) SELECT 'tenant',id,'operations','v2','v2',1 FROM tenant_row
	), domain_row AS (
	    INSERT INTO platform.domain_cutovers (
	    scope_kind,scope_id,domain,state,writer,generation,acceptance_digest,changed_at
) SELECT 'tenant',id,'workflow','v2','v2',1,NULL,pg_catalog.clock_timestamp()
  FROM tenant_row
)
INSERT INTO platform.cutover_group_members (
    scope_kind,scope_id,group_id,domain
) SELECT 'tenant',id,'operations','workflow' FROM tenant_row;`, tenantID)
	if err != nil {
		t.Fatalf("add isolated stale job tenant: %v", err)
	}
	return tenantID
}

func assertStaleJobRows(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	runningID uuid.UUID,
	queuedID uuid.UUID,
	effectID uuid.UUID,
) {
	t.Helper()
	for _, jobID := range []uuid.UUID{runningID, queuedID} {
		var state, code string
		var leaseOwner *uuid.UUID
		if err := fixture.owner.QueryRowContext(ctx, `
SELECT state, parked_code, lease_owner_id FROM platform.jobs WHERE id=$1`,
			jobID).Scan(&state, &code, &leaseOwner); err != nil {
			t.Fatal(err)
		}
		if state != "queued" || code != "JOB_AUTHORITY_STALE" || leaseOwner != nil {
			t.Fatalf("parked stale job %s = %s/%s/%v", jobID, state, code, leaseOwner)
		}
	}
	var roots, runningOpen, queuedAttempts int
	if err := fixture.owner.QueryRowContext(ctx, `
SELECT count(*) FROM platform.dead_letters
WHERE source_kind='job' AND source_id IN ($1,$2)
  AND reason_code='JOB_AUTHORITY_STALE'`, runningID, queuedID).Scan(&roots); err != nil {
		t.Fatal(err)
	}
	if err := fixture.owner.QueryRowContext(ctx, `
SELECT count(*) FILTER (WHERE job_id=$1 AND finished_at IS NULL),
       count(*) FILTER (WHERE job_id=$2)
FROM platform.job_attempts WHERE job_id IN ($1,$2)`,
		runningID, queuedID).Scan(&runningOpen, &queuedAttempts); err != nil {
		t.Fatal(err)
	}
	var effectState, effectCode string
	var effectOpen int
	if err := fixture.owner.QueryRowContext(ctx, `
SELECT job.state,job.last_error_code,
       count(*) FILTER (WHERE attempt.finished_at IS NULL)
FROM platform.jobs AS job
JOIN platform.job_attempts AS attempt ON attempt.job_id=job.id
WHERE job.id=$1 GROUP BY job.state,job.last_error_code`, effectID,
	).Scan(&effectState, &effectCode, &effectOpen); err != nil {
		t.Fatal(err)
	}
	if roots != 2 || runningOpen != 0 || queuedAttempts != 0 ||
		effectState != "dead" || effectCode != jobsv1.UnknownOutcomeCode ||
		effectOpen != 0 {
		t.Fatalf("stale evidence roots=%d open=%d queued=%d effect=%s/%s/%d",
			roots, runningOpen, queuedAttempts, effectState, effectCode, effectOpen)
	}
}
