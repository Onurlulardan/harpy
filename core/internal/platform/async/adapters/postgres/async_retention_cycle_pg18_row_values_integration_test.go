//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

func insertRetentionJobRows(
	t *testing.T, ctx context.Context, transaction *sql.Tx, ids retentionCycleWorkIDs,
) {
	t.Helper()
	if _, err := transaction.ExecContext(
		ctx, retentionTerminalJobsSeed, ids.redactJob, ids.purgeJob, asyncTenantA,
	); err != nil {
		t.Fatalf("seed terminal retention jobs: %v", err)
	}
	if _, err := transaction.ExecContext(
		ctx, retentionStaleJobSeed, ids.staleJob, asyncTenantA,
	); err != nil {
		t.Fatalf("seed stale retention job: %v", err)
	}
}

func insertRetentionReceiptRow(
	t *testing.T, ctx context.Context, transaction *sql.Tx, receiptID uuid.UUID,
) {
	t.Helper()
	if _, err := transaction.ExecContext(
		ctx, retentionReceiptSeed, receiptID, asyncTenantA,
	); err != nil {
		t.Fatalf("seed retention receipt: %v", err)
	}
}

func ageRetentionInboxRows(
	t *testing.T, ctx context.Context, transaction *sql.Tx, eventID uuid.UUID,
) {
	t.Helper()
	if _, err := transaction.ExecContext(ctx, `
UPDATE platform.inbox
SET recorded_at=pg_catalog.clock_timestamp()-interval '181 days'
WHERE event_id=$1 OR causation_id=$1`, eventID); err != nil {
		t.Fatalf("age retention inbox rows: %v", err)
	}
}

const retentionTerminalJobsSeed = `
WITH stamp AS MATERIALIZED (SELECT pg_catalog.clock_timestamp() AS now)
INSERT INTO platform.jobs (
 id,tenant_id,scope_kind,scope_id,authority_group_id,
 authority_group_generation,authority_domain,cutover_generation,writer_origin,
 kind,payload_revision,payload_ref,payload_sha256,payload_redacted_at,
 dedupe_key,effect_class,priority,run_at,state,attempts,max_attempts,
 lease_generation,retain_until,retention_release_revision,
 retention_release_sha256,retention_released_at,created_at,updated_at,completed_at
)
SELECT $1::uuid,$3::uuid,'tenant',$3::uuid,'operations',1,'workflow',1,'v2',
 'workflow.retention_redact',1,'payload/retention/redact',repeat('a',64),NULL,
 'retention-redact-'||$1::text,'database_transactional',0,now-'31 days'::interval,
 'succeeded',0,1,0,now+'100 days'::interval,0,NULL,NULL,
 now-'31 days'::interval,now-'31 days'::interval,now-'31 days'::interval
FROM stamp
UNION ALL
SELECT $2::uuid,$3::uuid,'tenant',$3::uuid,'operations',1,'workflow',1,'v2',
 'workflow.retention_purge',1,NULL,repeat('b',64),now-'350 days'::interval,
 'retention-purge-'||$2::text,'database_transactional',0,now-'400 days'::interval,
 'succeeded',0,1,0,now-'300 days'::interval,1,repeat('c',64),
 now-'1 day'::interval,now-'400 days'::interval,
 now-'1 day'::interval,now-'400 days'::interval
FROM stamp`

const retentionStaleJobSeed = `
WITH stamp AS MATERIALIZED (SELECT pg_catalog.clock_timestamp() AS now)
INSERT INTO platform.jobs (
 id,tenant_id,scope_kind,scope_id,authority_group_id,
 authority_group_generation,authority_domain,cutover_generation,writer_origin,
 kind,payload_revision,payload_ref,payload_sha256,dedupe_key,effect_class,
 priority,run_at,state,attempts,max_attempts,lease_generation,parked_at,
 parked_code,parked_current_group_generation,parked_current_cutover_generation,
 created_at,updated_at
)
SELECT $1::uuid,$2::uuid,'tenant',$2::uuid,'operations',1,'workflow',1,'v2',
 'workflow.retention_stale',1,'payload/retention/stale',repeat('d',64),
 'retention-stale-'||$1::text,'database_transactional',0,now-'2 days'::interval,
 'queued',0,1,0,now-'1 day'::interval,'JOB_AUTHORITY_STALE',2,2,
 now-'2 days'::interval,now-'1 day'::interval
FROM stamp`

const retentionReceiptSeed = `
WITH stamp AS MATERIALIZED (SELECT pg_catalog.clock_timestamp() AS now)
INSERT INTO platform.idempotency_receipts (
 receipt_id,record_kind,tenant_id,scope_kind,scope_id,authority_group_id,
 authority_group_generation,authority_domain,cutover_generation,writer_origin,
 request_sha256,retention_class,principal_id,operation_id,idempotency_key,
 authorization_version,response_status,response_headers,response_body,
 response_sha256,completed_at,retain_until
)
SELECT $1::uuid,'http',$2::uuid,'tenant',$2::uuid,'operations',1,'workflow',1,'v2',
 repeat('e',64),'interactive_write',$1::uuid,'workflow.retention','retention-receipt',
 1,200,'{}'::jsonb,pg_catalog.convert_to('{}','UTF8'),repeat('f',64),
 now-'2 days'::interval,now-'1 day'::interval
FROM stamp`
