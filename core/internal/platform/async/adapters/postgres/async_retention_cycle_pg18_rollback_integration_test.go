//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func assertRetentionCycleRollback(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	inboxEvent uuid.UUID,
) {
	t.Helper()
	jobID, inboxID := uuid.New(), uuid.New()
	seedRetentionRollbackRows(t, ctx, fixture, jobID, inboxID, inboxEvent)
	installRetentionFailureTrigger(t, ctx, fixture)

	var acquired bool
	var counts [7]int32
	err := fixture.runtime.QueryRow(ctx, retentionCycleCall).Scan(
		&acquired, &counts[0], &counts[1], &counts[2], &counts[3],
		&counts[4], &counts[5], &counts[6],
	)
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "P5799" {
		t.Fatalf("mid-cycle failure = %v", err)
	}
	var jobPreserved, inboxPreserved bool
	err = fixture.owner.QueryRowContext(ctx, `
SELECT
 (SELECT payload_ref IS NOT NULL AND payload_redacted_at IS NULL
    FROM platform.jobs WHERE id=$1),
 EXISTS (SELECT 1 FROM platform.inbox WHERE id=$2)`,
		jobID, inboxID,
	).Scan(&jobPreserved, &inboxPreserved)
	if err != nil || !jobPreserved || !inboxPreserved {
		t.Fatalf("retention rollback = job %t inbox %t: %v",
			jobPreserved, inboxPreserved, err)
	}
	removeRetentionFailureTrigger(t, ctx, fixture)
}

func seedRetentionRollbackRows(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	jobID, inboxID, inboxEvent uuid.UUID,
) {
	t.Helper()
	transaction, err := fixture.owner.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin retention rollback fixture: %v", err)
	}
	defer transaction.Rollback()
	for _, statement := range []string{
		"ALTER TABLE platform.jobs DISABLE TRIGGER ALL",
		"ALTER TABLE platform.inbox DISABLE TRIGGER ALL",
	} {
		if _, err = transaction.ExecContext(ctx, statement); err != nil {
			t.Fatalf("disable retention rollback triggers: %v", err)
		}
	}
	if _, err = transaction.ExecContext(
		ctx, retentionRollbackJobSeed, jobID, asyncTenantA,
	); err != nil {
		t.Fatalf("seed retention rollback job: %v", err)
	}
	if _, err = transaction.ExecContext(
		ctx, retentionRollbackInboxSeed, inboxID, inboxEvent,
	); err != nil {
		t.Fatalf("seed retention rollback inbox: %v", err)
	}
	for _, statement := range []string{
		"ALTER TABLE platform.inbox ENABLE TRIGGER ALL",
		"ALTER TABLE platform.jobs ENABLE TRIGGER ALL",
	} {
		if _, err = transaction.ExecContext(ctx, statement); err != nil {
			t.Fatalf("restore retention rollback triggers: %v", err)
		}
	}
	if err = transaction.Commit(); err != nil {
		t.Fatalf("commit retention rollback fixture: %v", err)
	}
}

const retentionRollbackJobSeed = `
WITH stamp AS MATERIALIZED (SELECT pg_catalog.clock_timestamp() AS now)
INSERT INTO platform.jobs (
 id,tenant_id,scope_kind,scope_id,authority_group_id,
 authority_group_generation,authority_domain,cutover_generation,writer_origin,
 kind,payload_revision,payload_ref,payload_sha256,dedupe_key,effect_class,
 priority,run_at,state,attempts,max_attempts,lease_generation,retain_until,
 created_at,updated_at,completed_at
)
SELECT $1::uuid,$2::uuid,'tenant',$2::uuid,'operations',1,'workflow',1,'v2',
 'workflow.retention_rollback',1,'payload/retention/rollback',repeat('1',64),
 'retention-rollback-'||$1::text,'database_transactional',0,
 now-'31 days'::interval,'succeeded',0,1,0,now+'100 days'::interval,
 now-'31 days'::interval,now-'31 days'::interval,now-'31 days'::interval
FROM stamp`

const retentionRollbackInboxSeed = `
INSERT INTO platform.inbox (
 id,record_kind,consumer_name,consumer_version,tenant_id,scope_kind,scope_id,
 producer_group_id,producer_group_generation,producer_domain,producer_generation,
 writer_origin,consumer_group_id,consumer_group_generation,consumer_domain,
 consumer_generation,consumer_writer_origin,aggregate_type,aggregate_id,
 event_id,event_type,event_version,event_sequence,receipt_sha256,recorded_at
)
SELECT $1::uuid,'event_receipt',consumer_name,consumer_version,tenant_id,scope_kind,
 scope_id,producer_group_id,producer_group_generation,producer_domain,
 producer_generation,writer_origin,consumer_group_id,consumer_group_generation,
 consumer_domain,consumer_generation,consumer_writer_origin,aggregate_type,
 aggregate_id,causation_id,'workflow.changed',1,through_sequence,repeat('2',64),
 pg_catalog.clock_timestamp()-interval '181 days'
FROM platform.inbox
WHERE record_kind='watermark' AND watermark_kind='event' AND causation_id=$2::uuid`

func installRetentionFailureTrigger(
	t *testing.T, ctx context.Context, fixture *eventGapPGFixture,
) {
	t.Helper()
	const statement = `
CREATE FUNCTION platform.retention_cycle_test_fail_delete()
RETURNS trigger LANGUAGE plpgsql AS $function$
BEGIN
 RAISE EXCEPTION USING ERRCODE='P5799',MESSAGE='retention cycle test failure';
END;
$function$;
CREATE TRIGGER retention_cycle_test_fail_delete
BEFORE DELETE ON platform.inbox FOR EACH ROW
EXECUTE FUNCTION platform.retention_cycle_test_fail_delete();`
	if _, err := fixture.owner.ExecContext(ctx, statement); err != nil {
		t.Fatalf("install retention failure trigger: %v", err)
	}
}

func removeRetentionFailureTrigger(
	t *testing.T, ctx context.Context, fixture *eventGapPGFixture,
) {
	t.Helper()
	if _, err := fixture.owner.ExecContext(ctx, `
DROP TRIGGER retention_cycle_test_fail_delete ON platform.inbox;
DROP FUNCTION platform.retention_cycle_test_fail_delete();`); err != nil {
		t.Fatalf("remove retention failure trigger: %v", err)
	}
}
