//go:build integration

package v1adoption

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func prepareNotificationCutoverPlatform(t *testing.T, ctx context.Context, db *pgx.Conn) {
	t.Helper()
	_, err := db.Exec(ctx, notificationCutoverPlatformSQL)
	if err != nil {
		t.Fatalf("prepare notification cutover platform: %v", err)
	}
}

const notificationCutoverPlatformSQL = `
CREATE TABLE platform.cutover_operations (
 operation_id uuid, run_id uuid, scope_kind text, scope_id uuid, group_id text,
 domain text, cutover_generation bigint, lease_manifest_sha256 text,
 state text, created_at timestamptz, PRIMARY KEY(operation_id,run_id)
);
CREATE TABLE platform.cutover_operation_checkpoint_keys (
 operation_id uuid, run_id uuid, relation_ordinal smallint, checkpoint_key text
);
CREATE TABLE platform.cutover_copy_checkpoints (
 operation_id uuid, run_id uuid, relation_ordinal smallint, checkpoint_key text,
 state text, final_lsn pg_lsn, PRIMARY KEY(operation_id,run_id,relation_ordinal,checkpoint_key)
);
CREATE TABLE platform.applied_change_receipts (
 receipt_id uuid, operation_id uuid, run_id uuid, manifest_sha256 text,
 source_xid xid8, source_commit_lsn pg_lsn, relation_ordinal smallint,
 change_ordinal integer, checkpoint_key text, mutation_kind text,
 mutation_sha256 text, applied_at timestamptz
);
CREATE TABLE platform.async_v1_adoption_runs (
 run_id uuid PRIMARY KEY, scope_kind text, scope_id uuid, domain text, group_id text,
 target_generation bigint, inventory_sha256 text, acceptance_sha256 text,
 item_count bigint, expected_enqueue_fences smallint,
 expected_scheduler_fences smallint, expected_consumer_fences smallint, state text
);
CREATE TABLE platform.async_v1_adoption_items (run_id uuid);
CREATE TABLE platform.async_v1_adoption_dispositions (
 run_id uuid, inventory_class text, source_name text, source_item_id text,
 receipt_sha256 text
);
CREATE TABLE platform.async_v1_restart_fence_receipts (
 run_id uuid, fence_kind text, component_name text, receipt_sha256 text
);
CREATE TABLE platform.async_v1_zero_inventory_receipts (
 run_id uuid, receipt_sha256 text, inventory_sha256 text,
 queue_count bigint, timer_count bigint, future_count bigint, running_count bigint
);
CREATE TABLE platform.contract_retirements (
 contract_key text, owner_module text, domain text, remaining_scope_count bigint,
 zero_hit_since timestamptz, state text, hit_count bigint, revision bigint
);
CREATE TABLE platform.notification_cutover_test_receipts (
 operation_id uuid, run_id uuid, relation_ordinal smallint, checkpoint_key text,
 batch_ordinal integer, batch_sha256 text, applied_count bigint,
 key_state_sha256 text, state_sha256 text,
 PRIMARY KEY(operation_id,run_id,relation_ordinal,checkpoint_key,batch_ordinal)
);
CREATE FUNCTION platform.cutover_prepare_batch_v1(
 p_operation_id uuid,p_run_id uuid,p_scope_kind text,p_scope_id uuid,
 p_relation_ordinal smallint,p_checkpoint_key text,p_batch_ordinal integer,
 p_source_commit_lsn pg_lsn,p_source_xid xid8,p_batch_sha256 text,
 p_lease_manifest_sha256 text
) RETURNS TABLE(
 procedure_contract_id text,procedure_oid bigint,procedure_signature text,
 already_applied boolean,applied_count bigint,key_state_sha256 text,state_sha256 text
) LANGUAGE plpgsql AS $fn$
DECLARE
 v_receipt platform.notification_cutover_test_receipts%ROWTYPE;
BEGIN
 SELECT receipt.* INTO v_receipt
 FROM platform.notification_cutover_test_receipts AS receipt
 WHERE receipt.operation_id=p_operation_id AND receipt.run_id=p_run_id
   AND receipt.relation_ordinal=p_relation_ordinal
   AND receipt.checkpoint_key=p_checkpoint_key
   AND receipt.batch_ordinal=p_batch_ordinal;
 IF FOUND AND v_receipt.batch_sha256 IS DISTINCT FROM p_batch_sha256 THEN
  RAISE EXCEPTION USING ERRCODE='55000',MESSAGE='test batch replay digest differs';
 END IF;
 RETURN QUERY SELECT 'notifications.v1_adoption.v1',
  'notifications_v1_adoption_apply_v1(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)'::regprocedure::oid::bigint,
  pg_catalog.format('%I.notifications_v1_adoption_apply_v1(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)',pg_catalog.current_schema()),
  FOUND,COALESCE(v_receipt.applied_count,0),
  COALESCE(v_receipt.key_state_sha256,pg_catalog.repeat('0',64)),
  COALESCE(v_receipt.state_sha256,pg_catalog.repeat('0',64));
END;
$fn$;
CREATE FUNCTION platform.cutover_record_batch_v1(
 p_operation_id uuid,p_run_id uuid,p_relation_ordinal smallint,
 p_checkpoint_key text,p_snapshot_name text,p_snapshot_lsn pg_lsn,
 p_received_lsn pg_lsn,p_commit_lsn pg_lsn,p_final_lsn pg_lsn,
 p_source_xid xid8,p_batch_ordinal integer,p_checkpoint_state text,
 p_last_key jsonb,p_batch_sha256 text,p_applied_bytes bigint,
 p_applied_count bigint,p_key_state_sha256 text,p_state_sha256 text
) RETURNS void LANGUAGE plpgsql AS $fn$
BEGIN
 INSERT INTO platform.notification_cutover_test_receipts VALUES(
  p_operation_id,p_run_id,p_relation_ordinal,p_checkpoint_key,p_batch_ordinal,
  p_batch_sha256,p_applied_count,p_key_state_sha256,p_state_sha256
 );
 UPDATE platform.cutover_copy_checkpoints SET state=p_checkpoint_state,
  final_lsn=p_final_lsn WHERE operation_id=p_operation_id AND run_id=p_run_id
  AND relation_ordinal=p_relation_ordinal AND checkpoint_key=p_checkpoint_key;
END;
$fn$;
GRANT USAGE ON SCHEMA platform TO harpy_adopter,harpy_runtime;
`
