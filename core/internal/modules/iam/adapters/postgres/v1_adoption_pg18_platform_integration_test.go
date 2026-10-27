//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const iamV1AdoptionCutoverStub = `
ALTER TABLE platform.cutover_groups
    ADD COLUMN state text NOT NULL DEFAULT 'shadow',
    ADD COLUMN writer text NOT NULL DEFAULT 'v1';
CREATE TABLE platform.test_iam_adoption_receipts (
    operation_id uuid NOT NULL, run_id uuid NOT NULL,
    relation_ordinal smallint NOT NULL, checkpoint_key text NOT NULL,
    batch_ordinal integer NOT NULL, source_xid xid8 NOT NULL,
    source_commit_lsn pg_lsn NOT NULL, batch_sha256 text NOT NULL,
    applied_count bigint NOT NULL, key_state_sha256 text NOT NULL,
    state_sha256 text NOT NULL,
    PRIMARY KEY (operation_id, run_id, relation_ordinal, checkpoint_key, batch_ordinal)
);
CREATE FUNCTION platform.cutover_prepare_batch_v1(
    p_operation_id uuid, p_run_id uuid, p_scope_kind text, p_scope_id uuid,
    p_relation_ordinal smallint, p_checkpoint_key text, p_batch_ordinal integer,
    p_source_commit_lsn pg_lsn, p_source_xid xid8, p_batch_sha256 text,
    p_lease_manifest_sha256 text
)
RETURNS TABLE (
    procedure_contract_id text, procedure_oid bigint, procedure_signature text,
    already_applied boolean, applied_count bigint,
    key_state_sha256 text, state_sha256 text
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_schema name;
    v_signature text;
    v_oid oid;
    v_found boolean;
    v_count bigint;
    v_key text;
    v_state text;
    v_sha text;
    v_lsn pg_lsn;
    v_xid xid8;
BEGIN
    IF SESSION_USER <> 'harpy_adopter' OR p_scope_kind <> 'tenant' OR
       p_lease_manifest_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION 'test cutover prepare denied';
    END IF;
    SELECT tenant.schema_name INTO STRICT v_schema
    FROM platform.tenants AS tenant WHERE tenant.id = p_scope_id;
    v_signature := pg_catalog.format(
        '%I.iam_v1_adoption_apply_v1(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)',
        v_schema
    );
    v_oid := pg_catalog.to_regprocedure(v_signature)::oid;
    SELECT receipt.applied_count, receipt.key_state_sha256, receipt.state_sha256,
        receipt.batch_sha256, receipt.source_commit_lsn, receipt.source_xid
    INTO v_count, v_key, v_state, v_sha, v_lsn, v_xid
    FROM platform.test_iam_adoption_receipts AS receipt
    WHERE receipt.operation_id = p_operation_id AND receipt.run_id = p_run_id
      AND receipt.relation_ordinal = p_relation_ordinal
      AND receipt.checkpoint_key = p_checkpoint_key
      AND receipt.batch_ordinal = p_batch_ordinal;
    v_found := FOUND;
    IF v_found AND (v_sha <> p_batch_sha256 OR v_lsn <> p_source_commit_lsn OR
                    v_xid <> p_source_xid) THEN
        RAISE EXCEPTION 'test cutover receipt differs';
    END IF;
    RETURN QUERY SELECT 'iam.v1_adoption.v1'::text, v_oid::bigint, v_signature,
        v_found, v_count, v_key, v_state;
END;
$function$;
CREATE FUNCTION platform.cutover_record_batch_v1(
    p_operation_id uuid, p_run_id uuid, p_relation_ordinal smallint,
    p_checkpoint_key text, p_snapshot_name text, p_snapshot_lsn pg_lsn,
    p_received_lsn pg_lsn, p_commit_lsn pg_lsn, p_final_lsn pg_lsn,
    p_source_xid xid8, p_batch_ordinal integer, p_checkpoint_state text,
    p_last_key jsonb, p_batch_sha256 text, p_applied_bytes bigint,
    p_applied_count bigint, p_key_state_sha256 text, p_state_sha256 text
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
BEGIN
    IF SESSION_USER <> 'harpy_adopter' OR p_checkpoint_state <> 'complete' OR
       p_snapshot_lsn > p_commit_lsn OR p_commit_lsn > p_received_lsn OR
       p_final_lsn IS NULL OR p_applied_bytes < 1 OR p_snapshot_name = '' OR
       (p_last_key IS NOT NULL AND
        pg_catalog.jsonb_typeof(p_last_key) NOT IN ('object', 'null')) THEN
        RAISE EXCEPTION 'test cutover record denied';
    END IF;
    INSERT INTO platform.test_iam_adoption_receipts (
        operation_id, run_id, relation_ordinal, checkpoint_key, batch_ordinal,
        source_xid, source_commit_lsn, batch_sha256, applied_count,
        key_state_sha256, state_sha256
    ) VALUES (
        p_operation_id, p_run_id, p_relation_ordinal, p_checkpoint_key,
        p_batch_ordinal, p_source_xid, p_commit_lsn, p_batch_sha256,
        p_applied_count, p_key_state_sha256, p_state_sha256
    );
END;
$function$;
`

func prepareIAMV1AdoptionCutoverStub(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, iamV1AdoptionCutoverStub); err != nil {
		t.Fatalf("prepare IAM V1 adoption cutover stub: %v", err)
	}
	if _, err := database.ExecContext(
		ctx, "GRANT USAGE ON SCHEMA platform TO harpy_adopter",
	); err != nil {
		t.Fatalf("grant IAM adopter platform routine access: %v", err)
	}
	for _, tenant := range iamFoundationTenants {
		statement := "GRANT USAGE ON SCHEMA " + tenant.schema + " TO harpy_adopter"
		if _, err := database.ExecContext(ctx, statement); err != nil {
			t.Fatalf("grant IAM adopter lease in %s: %v", tenant.schema, err)
		}
	}
}
