//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const (
	retentionSuccessJob    = "91000000-0000-4000-8000-000000000001"
	retentionTenantBJob    = "91000000-0000-4000-8000-000000000002"
	retentionRetryJob      = "91000000-0000-4000-8000-000000000003"
	retentionAttemptJob    = "91000000-0000-4000-8000-000000000004"
	retentionDependencyJob = "91000000-0000-4000-8000-000000000005"
	retentionEffectJob     = "91000000-0000-4000-8000-000000000006"
)

func seedRetentionReleaseFixture(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	if _, err := database.ExecContext(ctx, retentionReleaseFixtureSQL); err != nil {
		t.Fatalf("seed retention release fixture: %v", err)
	}
}

const retentionReleaseFixtureSQL = `
ALTER TABLE platform.jobs DISABLE TRIGGER ALL;
ALTER TABLE platform.job_attempts DISABLE TRIGGER ALL;
ALTER TABLE platform.idempotency_receipts DISABLE TRIGGER ALL;

INSERT INTO platform.jobs (
    id, tenant_id, scope_kind, scope_id, authority_group_id,
    authority_group_generation, authority_domain, cutover_generation,
    writer_origin, kind, payload_revision, payload_ref, payload_sha256,
    dedupe_key, effect_class, priority, run_at, state, attempts,
    max_attempts, lease_generation, retain_until, created_at, updated_at, completed_at
)
SELECT source.id, source.tenant_id, 'tenant', source.tenant_id, 'operations',
    1, 'workflow', 1, 'v2', 'workflow.execute', 1, 'payload/job/release',
    repeat('a', 64), source.id::text, 'database_transactional', 0,
    TIMESTAMPTZ '2025-01-01 00:00:00+00', 'succeeded', 0, 8, 0,
    TIMESTAMPTZ '2025-10-01 00:00:00+00',
    TIMESTAMPTZ '2025-01-01 00:00:00+00',
    TIMESTAMPTZ '2025-01-01 00:00:00+00',
    TIMESTAMPTZ '2025-01-01 00:00:00+00'
FROM (VALUES
    ('91000000-0000-4000-8000-000000000001'::uuid,
     '11111111-1111-4111-8111-111111111111'::uuid),
    ('91000000-0000-4000-8000-000000000002'::uuid,
     '22222222-2222-4222-8222-222222222222'::uuid),
    ('91000000-0000-4000-8000-000000000003'::uuid,
     '11111111-1111-4111-8111-111111111111'::uuid),
    ('91000000-0000-4000-8000-000000000004'::uuid,
     '11111111-1111-4111-8111-111111111111'::uuid),
    ('91000000-0000-4000-8000-000000000005'::uuid,
     '11111111-1111-4111-8111-111111111111'::uuid),
    ('91000000-0000-4000-8000-000000000006'::uuid,
     '11111111-1111-4111-8111-111111111111'::uuid)
) AS source(id, tenant_id);

INSERT INTO platform.jobs (
    id, tenant_id, scope_kind, scope_id, authority_group_id,
    authority_group_generation, authority_domain, cutover_generation,
    writer_origin, kind, payload_revision, payload_ref, payload_sha256,
    dedupe_key, effect_class, priority, run_at, state, attempts,
    max_attempts, lease_generation, retry_of_job_id
) VALUES (
    '92000000-0000-4000-8000-000000000001',
    '11111111-1111-4111-8111-111111111111', 'tenant',
    '11111111-1111-4111-8111-111111111111', 'operations', 1, 'workflow', 1,
    'v2', 'workflow.execute', 1, 'payload/job/retry', repeat('b', 64),
    'retention-retry-child', 'database_transactional', 0,
    pg_catalog.clock_timestamp(), 'queued', 0, 8, 0,
    '91000000-0000-4000-8000-000000000003'
);

INSERT INTO platform.job_attempts (
    attempt_id, job_id, scope_kind, scope_id, attempt,
    lease_generation, lease_owner_id, started_at
) VALUES (
    '93000000-0000-4000-8000-000000000001',
    '91000000-0000-4000-8000-000000000004', 'tenant',
    '11111111-1111-4111-8111-111111111111', 1, 1,
    '94000000-0000-4000-8000-000000000001', pg_catalog.clock_timestamp()
);

INSERT INTO platform.idempotency_receipts (
    receipt_id, record_kind, tenant_id, scope_kind, scope_id,
    authority_group_id, authority_group_generation, authority_domain,
    cutover_generation, writer_origin, request_sha256, retention_class,
    principal_id, operation_id, idempotency_key, authorization_version,
    response_status, response_headers, response_body, response_sha256,
    async_job_id, completed_at, retain_until
) VALUES (
    '95000000-0000-4000-8000-000000000001', 'http',
    '11111111-1111-4111-8111-111111111111', 'tenant',
    '11111111-1111-4111-8111-111111111111', 'operations', 1, 'workflow', 1,
    'v2', repeat('c', 64), 'async_handle',
    '96000000-0000-4000-8000-000000000001', 'workflow.execute',
    'retention-dependency', 1, 202, '{}'::jsonb,
    pg_catalog.convert_to('{}', 'UTF8'), repeat('d', 64),
    '91000000-0000-4000-8000-000000000005',
    pg_catalog.clock_timestamp(), pg_catalog.clock_timestamp() + interval '1 day'
);

INSERT INTO platform.idempotency_receipts (
    receipt_id, record_kind, tenant_id, scope_kind, scope_id,
    authority_group_id, authority_group_generation, authority_domain,
    cutover_generation, writer_origin, request_sha256, retention_class,
    effect_id, effect_kind, effect_revision, effect_key, provider, operation,
    correlation_id, causation_id, effect_status, effect_code,
    provider_reference, effect_result_ref, effect_result_sha256,
    completed_at, retain_until
) VALUES (
    '95000000-0000-4000-8000-000000000002', 'effect',
    '11111111-1111-4111-8111-111111111111', 'tenant',
    '11111111-1111-4111-8111-111111111111', 'operations', 1, 'workflow', 1,
    'v2', repeat('e', 64), 'external_effect',
    '97000000-0000-4000-8000-000000000001', 'workflow.effect', 1,
    'retention-effect', 'test.provider', 'execute', 'retention-effect',
    '98000000-0000-4000-8000-000000000001', 'succeeded', 'EFFECT_DONE',
    'provider-1', 'result/effect/1', repeat('f', 64),
    pg_catalog.clock_timestamp(), pg_catalog.clock_timestamp() + interval '1 day'
);

INSERT INTO platform.job_attempts (
    attempt_id, job_id, scope_kind, scope_id, attempt, lease_generation,
    lease_owner_id, started_at, finished_at, outcome,
    effect_started_at, effect_id, effect_request_sha256, effect_receipt_id
) VALUES (
    '93000000-0000-4000-8000-000000000002',
    '91000000-0000-4000-8000-000000000006', 'tenant',
    '11111111-1111-4111-8111-111111111111', 1, 1,
    '94000000-0000-4000-8000-000000000002',
    pg_catalog.clock_timestamp() - interval '1 hour',
    pg_catalog.clock_timestamp(), 'succeeded',
    pg_catalog.clock_timestamp() - interval '30 minutes',
    '97000000-0000-4000-8000-000000000001', repeat('e', 64),
    '95000000-0000-4000-8000-000000000002'
);

ALTER TABLE platform.idempotency_receipts ENABLE TRIGGER ALL;
ALTER TABLE platform.job_attempts ENABLE TRIGGER ALL;
ALTER TABLE platform.jobs ENABLE TRIGGER ALL;`
