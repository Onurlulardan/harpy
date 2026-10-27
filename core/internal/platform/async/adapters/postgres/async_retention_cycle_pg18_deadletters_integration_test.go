//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func seedRetentionDeadLetters(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	ids retentionCycleWorkIDs,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin retention dead-letter fixture: %v", err)
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(
		ctx, "ALTER TABLE platform.dead_letters DISABLE TRIGGER ALL",
	); err != nil {
		t.Fatalf("disable retention dead-letter triggers: %v", err)
	}
	if _, err = transaction.ExecContext(
		ctx, retentionDeadLetterRowsSeed, ids.purgeRoot, ids.staleRoot,
		ids.staleJob, ids.liveRoot,
	); err != nil {
		t.Fatalf("seed retention dead letters: %v", err)
	}
	if _, err = transaction.ExecContext(
		ctx, "ALTER TABLE platform.dead_letters ENABLE TRIGGER ALL",
	); err != nil {
		t.Fatalf("restore retention dead-letter triggers: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatalf("commit retention dead-letter fixture: %v", err)
	}
}

const retentionDeadLetterRowsSeed = `
WITH stamp AS MATERIALIZED (SELECT pg_catalog.clock_timestamp() AS now)
INSERT INTO platform.dead_letters (
 id,root_id,entry_sequence,record_kind,tenant_id,scope_kind,scope_id,
 authority_group_id,authority_group_generation,authority_domain,
 cutover_generation,writer_origin,source_kind,source_id,consumer_name,
 consumer_version,source_contract,source_version,source_sha256,
 reason_class,reason_code,redacted_error_sha256,payload_canonical,
 payload_sha256,payload_redacted_at,redacted_envelope,
 full_payload_expires_at,metadata_expires_at,recorded_at
)
SELECT $1::uuid,$1::uuid,0,'root',root.tenant_id,root.scope_kind,root.scope_id,
 root.authority_group_id,root.authority_group_generation,root.authority_domain,
 root.cutover_generation,'v2','event',$1::uuid,root.consumer_name,
 root.consumer_version,root.source_contract,root.source_version,root.source_sha256,
 'permanent','EVENT_PERMANENT_FAILURE',root.redacted_error_sha256,NULL,
 root.payload_sha256,stamp.now-interval '2 days',root.redacted_envelope,
 stamp.now-interval '3 days',stamp.now-interval '1 day',
 stamp.now-interval '4 days'
FROM platform.dead_letters AS root CROSS JOIN stamp WHERE root.id=$4::uuid
UNION ALL
SELECT $2::uuid,$2::uuid,0,'root',root.tenant_id,root.scope_kind,root.scope_id,
 root.authority_group_id,root.authority_group_generation,root.authority_domain,
 root.cutover_generation,'v2','job',$3::uuid,NULL,NULL,
 'workflow.retention_stale',1,repeat('d',64),
 'stale_generation','JOB_AUTHORITY_STALE',root.redacted_error_sha256,
 root.payload_canonical,root.payload_sha256,NULL,root.redacted_envelope,
 stamp.now-interval '1 day',stamp.now+interval '365 days',
 stamp.now-interval '2 days'
FROM platform.dead_letters AS root CROSS JOIN stamp WHERE root.id=$4::uuid`
