-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_v1_adoption_apply_v1(
    p_operation_id uuid,
    p_run_id uuid,
    p_scope_kind text,
    p_scope_id uuid,
    p_relation_ordinal smallint,
    p_checkpoint_key text,
    p_snapshot_lsn pg_lsn,
    p_received_lsn pg_lsn,
    p_commit_lsn pg_lsn,
    p_source_xid xid8,
    p_batch_ordinal integer,
    p_change_batch jsonb,
    p_batch_sha256 text
)
RETURNS TABLE (applied_count bigint, key_state_sha256 text, state_sha256 text)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_change jsonb;
    v_payload jsonb;
    v_rows bigint;
    v_applied bigint;
    v_key_sha text;
    v_state_sha text;
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    v_applied := 0;
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_kind IS DISTINCT FROM 'tenant' OR
       p_batch_sha256 !~ '^[0-9a-f]{64}$' OR NOT EXISTS (
           SELECT 1 FROM iam_v1_adoption_runs AS run
           WHERE run.operation_id = p_operation_id AND run.run_id = p_run_id
             AND run.tenant_id = p_scope_id AND run.eligible
             AND run.state IN ('planned', 'copying')
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM V1 adoption apply authority or plan denied';
    END IF;
    FOR v_change IN
        SELECT value FROM pg_catalog.jsonb_array_elements(p_change_batch->'changes')
            WITH ORDINALITY AS item(value, position)
        ORDER BY position
    LOOP
        v_payload := COALESCE(v_change->'after', v_change->'before');
        INSERT INTO iam_v1_adoption_stage (
            operation_id, run_id, source_category, source_key,
            target_table, target_key, target_record, record_sha256,
            removed, source_xid, batch_ordinal, updated_at
        ) VALUES (
            p_operation_id, p_run_id, v_payload->>'source_category',
            v_payload->>'source_key', v_payload->>'target_table',
            v_payload->>'target_key', v_payload->'target_record',
            pg_catalog.sha256(pg_catalog.convert_to(
                (v_payload->'target_record')::text, 'UTF8'
            )),
            v_change->>'kind' = 'delete', p_source_xid, p_batch_ordinal, v_now
        )
        ON CONFLICT (operation_id, run_id, target_table, target_key) DO UPDATE
        SET source_category = EXCLUDED.source_category,
            source_key = EXCLUDED.source_key,
            target_record = EXCLUDED.target_record,
            record_sha256 = EXCLUDED.record_sha256,
            removed = EXCLUDED.removed,
            source_xid = EXCLUDED.source_xid,
            batch_ordinal = EXCLUDED.batch_ordinal,
            updated_at = EXCLUDED.updated_at
        WHERE iam_v1_adoption_stage.batch_ordinal < EXCLUDED.batch_ordinal;
        GET DIAGNOSTICS v_rows = ROW_COUNT;
        v_applied := v_applied + v_rows;
    END LOOP;
    SELECT pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        COALESCE(pg_catalog.string_agg(
            stage.target_table || ':' || stage.target_key || ':' ||
            pg_catalog.encode(stage.record_sha256, 'hex') || ':' ||
            stage.removed::text, ',' ORDER BY stage.target_table, stage.target_key
        ), ''), 'UTF8'
    )), 'hex'), pg_catalog.count(*)
    INTO v_key_sha, v_rows
    FROM iam_v1_adoption_stage AS stage
    WHERE stage.operation_id = p_operation_id AND stage.run_id = p_run_id
      AND stage.source_category = p_checkpoint_key AND NOT stage.removed;
    UPDATE iam_v1_adoption_category_evidence
    SET staged_count = v_rows, staged_sha256 = pg_catalog.decode(v_key_sha, 'hex'),
        relation_ordinal = p_relation_ordinal, copy_state = p_change_batch->>'checkpoint_state',
        last_batch_ordinal = p_batch_ordinal, updated_at = v_now
    WHERE operation_id = p_operation_id AND run_id = p_run_id
      AND category = p_checkpoint_key
      AND (relation_ordinal IS NULL OR relation_ordinal = p_relation_ordinal)
      AND (last_batch_ordinal IS NULL OR last_batch_ordinal < p_batch_ordinal)
      AND CASE copy_state
          WHEN 'planned' THEN p_change_batch->>'checkpoint_state'
              IN ('initial_copy', 'delta', 'complete')
          WHEN 'initial_copy' THEN p_change_batch->>'checkpoint_state'
              IN ('delta', 'complete')
          WHEN 'delta' THEN p_change_batch->>'checkpoint_state'
              IN ('delta', 'complete')
          ELSE false
      END;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '22000',
            MESSAGE = 'IAM V1 adoption category checkpoint regressed';
    END IF;
    UPDATE iam_v1_adoption_runs
    SET state = 'copying', updated_at = v_now
    WHERE operation_id = p_operation_id AND run_id = p_run_id
      AND state IN ('planned', 'copying');
    SELECT pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        COALESCE(pg_catalog.string_agg(
            stage.target_table || ':' || stage.target_key || ':' ||
            pg_catalog.encode(stage.record_sha256, 'hex') || ':' ||
            stage.removed::text, ',' ORDER BY stage.target_table, stage.target_key
        ), ''), 'UTF8'
    )), 'hex')
    INTO v_state_sha
    FROM iam_v1_adoption_stage AS stage
    WHERE stage.operation_id = p_operation_id AND stage.run_id = p_run_id
      AND NOT stage.removed;
    RETURN QUERY SELECT v_applied, v_key_sha, v_state_sha;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_v1_adoption_apply_v1(
    uuid, uuid, text, uuid, smallint, text, pg_lsn, pg_lsn, pg_lsn,
    xid8, integer, jsonb, text
) FROM PUBLIC;
