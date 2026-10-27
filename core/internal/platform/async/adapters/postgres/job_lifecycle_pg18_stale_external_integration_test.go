//go:build integration

package postgres

import (
	"context"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

func assertStaleExternalUnknownRace(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
) {
	t.Helper()
	tenantID := uuid.MustParse(asyncTenantB)
	authority, err := newJobAuthority(
		fixture.store.identity, "tenant", tenantID,
		"operations", 1, "workflow", 1,
	)
	if err != nil {
		t.Fatal(err)
	}
	fixture.authority = authority
	assertStaleExternalFinalizerACL(t, ctx, fixture)

	currentID := insertQueuedLifecycleJob(
		t, ctx, fixture, asyncTenantB, jobsv1.EffectExternal, 8,
	)
	currentClaim := claimOneExternalJob(t, ctx, fixture, currentID)
	currentPermit := markExternalEffect(t, ctx, fixture, currentClaim)
	currentWork := staleExternalFromPermit(t, fixture, currentPermit)
	if finalized, err := fixture.store.finalizeStaleExternalJob(
		ctx, currentWork,
	); err != nil || finalized {
		t.Fatalf("current authority finalization = %t, %v", finalized, err)
	}
	state, open := lifecycleState(t, ctx, fixture, currentID)
	if state != "running" || open != 1 {
		t.Fatalf("current authority job changed: %s/%d", state, open)
	}
	if _, err := fixture.store.MarkExternalJobOutcomeUnknown(
		ctx, currentPermit,
	); err != nil {
		t.Fatalf("clean up current authority job: %v", err)
	}

	raceID := insertQueuedLifecycleJob(
		t, ctx, fixture, asyncTenantB, jobsv1.EffectExternal, 8,
	)
	raceClaim := claimOneExternalJob(t, ctx, fixture, raceID)
	racePermit := markExternalEffect(t, ctx, fixture, raceClaim)
	raceWork := staleExternalFromPermit(t, fixture, racePermit)
	setPublisherDomainGeneration(t, ctx, fixture.publisherPGFixture, tenantID, 2)

	type result struct {
		finalized bool
		err       error
	}
	results := make(chan result, 2)
	for range 2 {
		go func() {
			finalized, err := fixture.store.finalizeStaleExternalJob(ctx, raceWork)
			results <- result{finalized: finalized, err: err}
		}()
	}
	finalized := 0
	for range 2 {
		result := <-results
		if result.err != nil {
			t.Fatalf("race finalizer: %v", result.err)
		}
		if result.finalized {
			finalized++
		}
	}
	if finalized != 1 {
		t.Fatalf("race winners = %d, want 1", finalized)
	}
	assertStaleExternalTerminal(t, ctx, fixture, raceID)
	if restarted, err := fixture.store.finalizeStaleExternalJob(
		ctx, raceWork,
	); err != nil || restarted {
		t.Fatalf("restart stale external finalization = %t, %v", restarted, err)
	}
}

func assertStaleExternalFinalizerACL(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
) {
	t.Helper()
	var securityDefiner, safePath, runtimeCan, adopterCan, adminCan bool
	var owner string
	err := fixture.owner.QueryRowContext(ctx, `
SELECT routine.prosecdef,owner.rolname,
       routine.proconfig=ARRAY['search_path=pg_catalog']::text[],
       has_function_privilege('harpy_runtime',routine.oid,'EXECUTE'),
       has_function_privilege('harpy_adopter',routine.oid,'EXECUTE'),
       has_function_privilege('harpy_cluster_admin',routine.oid,'EXECUTE')
FROM pg_catalog.pg_proc AS routine
JOIN pg_catalog.pg_roles AS owner ON owner.oid=routine.proowner
WHERE routine.oid=pg_catalog.to_regprocedure(
 'platform.platform_async_finalize_stale_external_unknown_v1(uuid,uuid,bigint,uuid,timestamptz)'
)`).Scan(&securityDefiner, &owner, &safePath, &runtimeCan, &adopterCan, &adminCan)
	if err != nil {
		t.Fatal(err)
	}
	if !securityDefiner || owner != "harpy_owner" || !safePath ||
		!runtimeCan || adopterCan || adminCan {
		t.Fatalf("stale finalizer ACL=%t/%s/%t/%t/%t/%t",
			securityDefiner, owner, safePath, runtimeCan, adopterCan, adminCan)
	}
}

func staleExternalFromPermit(
	t *testing.T,
	fixture jobLifecyclePGFixture,
	permit ExternalEffectPermit,
) staleExternalJob {
	t.Helper()
	claim := permit.claim
	job, err := newStaleExternalJob(
		fixture.store.identity, claim.JobID(), claim.LeaseOwnerID(),
		claim.LeaseGeneration(), claim.AttemptID(), permit.StartedAt(),
	)
	if err != nil {
		t.Fatal(err)
	}
	return job
}

func assertStaleExternalTerminal(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	jobID uuid.UUID,
) {
	t.Helper()
	var state, jobCode, outcome, attemptCode string
	var receiptCount, roots int
	err := fixture.owner.QueryRowContext(ctx, `
SELECT job.state,job.last_error_code,attempt.outcome,attempt.error_code,
       (SELECT count(*) FROM platform.idempotency_receipts WHERE effect_id=attempt.effect_id),
       (SELECT count(*) FROM platform.dead_letters
        WHERE source_kind='job' AND source_id=job.id
          AND reason_class='outcome_unknown' AND reason_code='JOB_OUTCOME_UNKNOWN')
FROM platform.jobs AS job
JOIN platform.job_attempts AS attempt ON attempt.job_id=job.id
WHERE job.id=$1`, jobID).Scan(
		&state, &jobCode, &outcome, &attemptCode, &receiptCount, &roots,
	)
	if err != nil {
		t.Fatal(err)
	}
	if state != "dead" || jobCode != jobsv1.UnknownOutcomeCode ||
		outcome != "dead" || attemptCode != jobsv1.UnknownOutcomeCode ||
		receiptCount != 0 || roots != 1 {
		t.Fatalf("stale external terminal=%s/%s %s/%s receipts=%d roots=%d",
			state, jobCode, outcome, attemptCode, receiptCount, roots)
	}
}
