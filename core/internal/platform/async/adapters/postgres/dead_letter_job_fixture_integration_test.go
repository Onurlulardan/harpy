//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func seedDeadJobRoot(
	t *testing.T,
	ctx context.Context,
	fixture deadLetterPGFixture,
	tenantID uuid.UUID,
) (uuid.UUID, uuid.UUID) {
	t.Helper()
	jobID, workerID, attemptID := uuid.New(), uuid.New(), uuid.New()
	jobPayloadSHA := deadLetterDigest([]byte("job-payload"))
	transaction, err := fixture.runtime.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback(context.Background())
	_, err = transaction.Exec(ctx, `
WITH job_clock AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS now
)
INSERT INTO platform.jobs (
    id, tenant_id, scope_kind, scope_id,
    authority_group_id, authority_group_generation,
    authority_domain, cutover_generation, writer_origin,
    kind, payload_revision, payload_ref, payload_sha256, dedupe_key,
    effect_class, effect_idempotency_key, priority, run_at,
    state, attempts, max_attempts, lease_owner_id, lease_generation,
    lease_heartbeat_at, lease_until, created_at, updated_at
)
SELECT
    $1, $2, 'tenant', $2,
    'operations', 1, 'workflow', 1, 'v2',
    'workflow.test_job', 1, 'jobs/source/42', $3, 'source-job-42',
    'external', 'provider-effect/42', 7, now - interval '2 days',
    'running', 8, 8, $4, 8,
    now - interval '2 seconds', now + interval '58 seconds',
    now - interval '2 days', now - interval '2 seconds'
FROM job_clock`, jobID, tenantID, jobPayloadSHA, workerID)
	if err != nil {
		t.Fatalf("seed dead source job: %v", err)
	}
	_, err = transaction.Exec(ctx, `
INSERT INTO platform.job_attempts (
    attempt_id,job_id,scope_kind,scope_id,attempt,lease_generation,
    lease_owner_id,started_at
)
SELECT $1,id,scope_kind,scope_id,attempts,lease_generation,
       lease_owner_id,lease_heartbeat_at
FROM platform.jobs WHERE id=$2`, attemptID, jobID)
	if err != nil {
		t.Fatalf("seed dead source attempt: %v", err)
	}
	errorSHA := deadLetterDigest([]byte("error"))
	_, err = transaction.Exec(ctx, `
WITH terminal_clock AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS now
)
UPDATE platform.job_attempts AS attempt
SET finished_at=terminal_clock.now,outcome='dead',error_class='permanent',
    error_code='PROVIDER_REJECTED',redacted_error_sha256=$2
FROM terminal_clock
WHERE attempt.attempt_id=$1`, attemptID, errorSHA)
	if err != nil {
		t.Fatalf("terminalize source attempt: %v", err)
	}
	_, err = transaction.Exec(ctx, `
WITH terminal_clock AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS now
)
UPDATE platform.jobs AS job
SET state='dead',lease_owner_id=NULL,lease_heartbeat_at=NULL,lease_until=NULL,
    last_error_class='permanent',last_error_code='PROVIDER_REJECTED',
    last_error_sha256=$2,retain_until=terminal_clock.now+interval '365 days',
    updated_at=terminal_clock.now,completed_at=terminal_clock.now
FROM terminal_clock WHERE job.id=$1`, jobID, errorSHA)
	if err != nil {
		t.Fatalf("terminalize source job: %v", err)
	}
	var rootID uuid.UUID
	if err := transaction.QueryRow(ctx, `
SELECT id FROM platform.dead_letters
WHERE source_kind='job' AND record_kind='root' AND source_id=$1`, jobID).Scan(&rootID); err != nil {
		t.Fatalf("read automatic job dead-letter root: %v", err)
	}
	if err := transaction.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	return jobID, rootID
}

func seedExpiredEventRoot(
	t *testing.T,
	ctx context.Context,
	fixture deadLetterPGFixture,
	tenantID uuid.UUID,
) uuid.UUID {
	t.Helper()
	rootID := uuid.New()
	_, err := fixture.runtime.Exec(ctx, `
WITH dead_clock AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS now
)
INSERT INTO platform.dead_letters (
    id, root_id, entry_sequence, record_kind,
    tenant_id, scope_kind, scope_id,
    authority_group_id, authority_group_generation,
    authority_domain, cutover_generation, writer_origin,
    source_kind, source_id, consumer_name, consumer_version,
    source_contract, source_version, source_sha256,
    reason_class, reason_code, redacted_error_sha256,
    payload_canonical, payload_sha256, payload_redacted_at,
    redacted_envelope,
    full_payload_expires_at, metadata_expires_at, recorded_at
)
SELECT
    $1, $1, 0, 'root',
    $2, 'tenant', $2,
    'operations', 1, 'workflow', 1, 'v2',
    'event', $3, 'workflow.projector', 1,
    'forms.entry.created', 1, $4,
    'delivery_limit', 'EVENT_DELIVERY_LIMIT', $5,
    NULL, $6, now,
    '{"event_id":"expired","business_effect_idempotency_key":"expired"}'::jsonb,
    now - interval '1 day', now + interval '300 days', now - interval '31 days'
FROM dead_clock`, rootID, tenantID, uuid.New(), deadLetterDigest([]byte("source")),
		deadLetterDigest([]byte("error")), deadLetterDigest([]byte("payload")))
	if err != nil {
		t.Fatalf("seed expired dead-letter root: %v", err)
	}
	return rootID
}
