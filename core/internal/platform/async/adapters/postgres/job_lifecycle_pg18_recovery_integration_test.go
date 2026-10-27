//go:build integration

package postgres

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

func assertExpiredJobRecoveryRace(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
) {
	t.Helper()
	requeuedID := insertExpiredLifecycleJob(
		t, ctx, fixture, jobsv1.EffectDatabaseTransactional, 1, 8, false,
	)
	results := make(chan JobRecoveryResult, 2)
	errors := make(chan error, 2)
	var wait sync.WaitGroup
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, err := fixture.store.RecoverExpiredJobs(ctx, fixture.authority, 1)
			results <- result
			errors <- err
		}()
	}
	wait.Wait()
	close(results)
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	total := 0
	for result := range results {
		total += result.Total()
	}
	if total != 1 {
		t.Fatalf("recovery race terminalized %d attempts", total)
	}
	state, open := lifecycleState(t, ctx, fixture, requeuedID)
	if state != string(jobsv1.StateQueued) || open != 0 {
		t.Fatalf("recovered retry state=%s open=%d", state, open)
	}

	exhaustedID := insertExpiredLifecycleJob(
		t, ctx, fixture, jobsv1.EffectDatabaseTransactional, 8, 8, false,
	)
	result, err := fixture.store.RecoverExpiredJobs(ctx, fixture.authority, 10)
	if err != nil || result.Dead() != 1 {
		t.Fatalf("max attempt recovery=%#v error=%v", result, err)
	}
	state, open = lifecycleState(t, ctx, fixture, exhaustedID)
	if state != string(jobsv1.StateDead) || open != 0 {
		t.Fatalf("exhausted recovery state=%s open=%d", state, open)
	}
	assertTerminalJobRoot(
		t, ctx, fixture, exhaustedID, "delivery_limit", "JOB_ATTEMPTS_EXHAUSTED",
	)

	unknownID := insertExpiredLifecycleJob(
		t, ctx, fixture, jobsv1.EffectExternal, 1, 8, true,
	)
	result, err = fixture.store.RecoverExpiredJobs(ctx, fixture.authority, 10)
	if err != nil || result.Dead() != 1 || result.Requeued() != 0 {
		t.Fatalf("external crash recovery=%#v error=%v", result, err)
	}
	state, open = lifecycleState(t, ctx, fixture, unknownID)
	if state != string(jobsv1.StateDead) || open != 0 {
		t.Fatalf("unknown recovery state=%s open=%d", state, open)
	}
	assertTerminalJobRoot(
		t, ctx, fixture, unknownID, "outcome_unknown", jobsv1.UnknownOutcomeCode,
	)
}

func insertExpiredLifecycleJob(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	effect jobsv1.EffectClass,
	attempt int,
	maximumAttempts int,
	effectStarted bool,
) uuid.UUID {
	t.Helper()
	jobID, workerID := uuid.New(), uuid.New()
	var effectKey any
	var effectStartedSQL any
	if effect == jobsv1.EffectExternal {
		effectKey = "effect:expired:" + jobID.String()
	}
	if effectStarted {
		effectStartedSQL = time.Now().UTC().Add(-2 * time.Minute)
	}
	const statement = `
INSERT INTO platform.jobs (
 id,tenant_id,scope_kind,scope_id,authority_group_id,authority_group_generation,
 authority_domain,cutover_generation,writer_origin,kind,payload_revision,payload_ref,
 payload_sha256,dedupe_key,effect_class,effect_idempotency_key,priority,run_at,state,
 attempts,max_attempts,lease_owner_id,lease_generation,lease_heartbeat_at,lease_until,
 effect_started_at,created_at,updated_at
) VALUES (
 $1,$2,'tenant',$2,'operations',1,'workflow',1,'v2','workflow.execute',1,
 $3,repeat('a',64),$4,$5,$6,1,pg_catalog.clock_timestamp()-interval '4 minutes',
 'running',$7::integer,$8,$9,$7::bigint,pg_catalog.clock_timestamp()-interval '3 minutes',
 pg_catalog.clock_timestamp()-interval '2 minutes',$10,
 pg_catalog.clock_timestamp()-interval '5 minutes',
 pg_catalog.clock_timestamp()-interval '3 minutes'
)`
	_, err := fixture.runtime.Exec(
		ctx, statement, jobID, asyncTenantA, "payload/"+jobID.String(),
		"dedupe:"+jobID.String(), string(effect), effectKey,
		attempt, maximumAttempts, workerID, effectStartedSQL,
	)
	if err != nil {
		t.Fatalf("insert expired job: %v", err)
	}
	var attemptID uuid.UUID
	err = fixture.runtime.QueryRow(ctx, `
INSERT INTO platform.job_attempts (
 attempt_id,job_id,scope_kind,scope_id,attempt,lease_generation,lease_owner_id,started_at
) SELECT pg_catalog.uuidv7(),id,scope_kind,scope_id,attempts,lease_generation,
 lease_owner_id,lease_heartbeat_at FROM platform.jobs WHERE id=$1
RETURNING attempt_id`, jobID).Scan(&attemptID)
	if err != nil {
		t.Fatalf("insert expired attempt: %v", err)
	}
	if effectStarted {
		result, err := fixture.runtime.Exec(
			ctx, markExternalAttemptEffectSQL, attemptID, jobID, workerID, int64(attempt),
			effectStartedSQL, uuid.New(), strings.Repeat("d", 64),
		)
		if err != nil || result.RowsAffected() != 1 {
			t.Fatalf("mark expired external attempt: %v %d", err, result.RowsAffected())
		}
	}
	return jobID
}
