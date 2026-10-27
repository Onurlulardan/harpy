//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func runSchemaExecutorLeaseRecovery(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tenantID string,
) {
	formID, jobID := schemaExecutorIntegrationUUID(300), schemaExecutorIntegrationUUID(301)
	environment.seedPlan(t, ctx, jobID, schemaExecutorCreatePlan(t, tenantID, formID, 1))
	record, found, err := environment.store.Claim(ctx, schemaexecutor.ClaimRequest{
		ExecutorID: schemaExecutorIntegrationUUID(302),
		LeaseID:    schemaExecutorIntegrationUUID(303),
	})
	if err != nil || !found {
		t.Fatalf("claim recovery job: found=%t err=%v", found, err)
	}
	lease, _, err := schemaexecutor.AcceptIntegrationClaim(record)
	if err != nil {
		t.Fatal(err)
	}
	staleLease, _, err := schemaexecutor.AcceptIntegrationClaim(record)
	if err != nil {
		t.Fatal(err)
	}
	if err := environment.store.MarkRunning(ctx, lease); err != nil {
		t.Fatal(err)
	}
	if err := environment.store.Heartbeat(
		ctx, schemaexecutor.HeartbeatRequest{Lease: staleLease},
	); err == nil {
		t.Fatal("expected stale generation heartbeat rejection")
	}
	if err := environment.store.Heartbeat(
		ctx, schemaexecutor.HeartbeatRequest{Lease: lease},
	); err != nil {
		t.Fatalf("heartbeat active lease: %v", err)
	}
	expireSchemaExecutorJob(t, ctx, environment, jobID)
	recovered, err := environment.store.RecoverExpired(ctx)
	if err != nil || !recovered {
		t.Fatalf("recover expired lease: recovered=%t err=%v", recovered, err)
	}
	var state string
	if err := environment.pool.QueryRow(ctx,
		"SELECT state FROM platform.schema_jobs WHERE job_id = $1", jobID,
	).Scan(&state); err != nil || state != "queued" {
		t.Fatalf("recovered state=%s err=%v", state, err)
	}
	_, activeLease, activePlan := environment.startPlan(t, ctx,
		schemaExecutorIntegrationUUID(304), schemaExecutorIntegrationUUID(305))
	if err := environment.store.Apply(ctx, schemaexecutor.ApplyRequest{
		Lease: activeLease, Plan: activePlan,
	}); err != nil {
		t.Fatalf("complete recovered job: %v", err)
	}
	runSchemaExecutorAttemptExhaustion(t, ctx, environment, tenantID)
}

func runSchemaExecutorAttemptExhaustion(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tenantID string,
) {
	jobID := schemaExecutorIntegrationUUID(310)
	environment.seedPlan(t, ctx, jobID,
		schemaExecutorCreatePlan(t, tenantID, schemaExecutorIntegrationUUID(311), 1))
	for attempt := 1; attempt <= schemaexecutor.MaximumAttempts; attempt++ {
		record, _, _ := environment.startPlan(t, ctx,
			schemaExecutorIntegrationUUID(320+attempt),
			schemaExecutorIntegrationUUID(340+attempt),
		)
		if record.Lease.JobAttempt != attempt || record.Lease.StepAttempt != attempt {
			t.Fatalf("attempt %d claim job=%d step=%d", attempt,
				record.Lease.JobAttempt, record.Lease.StepAttempt)
		}
		if record.Job.JobID != jobID {
			t.Fatalf("attempt %d claimed job %s, want %s", attempt, record.Job.JobID, jobID)
		}
		expireSchemaExecutorJob(t, ctx, environment, jobID)
		recovered, err := environment.store.RecoverExpired(ctx)
		if err != nil || !recovered {
			t.Fatalf("recover attempt %d: recovered=%t err=%v", attempt, recovered, err)
		}
	}
	var state, code string
	if err := environment.pool.QueryRow(ctx, `
SELECT state, failure_code FROM platform.schema_jobs WHERE job_id = $1`, jobID,
	).Scan(&state, &code); err != nil || state != "failed" || code != "attempt_limit_exhausted" {
		t.Fatalf("exhausted job state=%s code=%s err=%v", state, code, err)
	}
}

func expireSchemaExecutorJob(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	jobID string,
) {
	t.Helper()
	if _, err := environment.pool.Exec(
		ctx, "ALTER TABLE platform.schema_jobs DISABLE TRIGGER USER",
	); err != nil {
		t.Fatal(err)
	}
	_, updateErr := environment.pool.Exec(ctx, `
WITH authoritative_time AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS now
)
UPDATE platform.schema_jobs
SET created_at = authoritative_time.now - INTERVAL '5 minutes',
    available_at = authoritative_time.now - INTERVAL '5 minutes',
    leased_at = authoritative_time.now - INTERVAL '4 minutes',
    lease_heartbeat_at = authoritative_time.now - INTERVAL '3 minutes',
    updated_at = authoritative_time.now - INTERVAL '2 minutes',
    lease_expires_at = authoritative_time.now - INTERVAL '1 minute',
    last_lease_expired_at = CASE
        WHEN last_lease_expired_at IS NULL THEN NULL
        ELSE authoritative_time.now - INTERVAL '2 minutes'
    END
FROM authoritative_time
WHERE job_id = $1`, jobID)
	_, enableErr := environment.pool.Exec(
		ctx, "ALTER TABLE platform.schema_jobs ENABLE TRIGGER USER",
	)
	if updateErr != nil {
		t.Fatalf("expire schema job: %v", updateErr)
	}
	if enableErr != nil {
		t.Fatalf("restore schema job triggers: %v", enableErr)
	}
}
