-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_directory_repair_finish_v1(
    p_scope_schema name, p_tenant_id uuid, p_job_id uuid,
    p_attempt_id uuid, p_payload_revision integer,
    p_payload_reference text, p_payload_sha256 bytea,
    p_cursor_id uuid, p_scanned_count bigint, p_affected_count bigint,
    p_result_sha256 bytea, p_successor_job_id uuid
)
RETURNS TABLE (
    successor_job_id uuid, successor_run_at timestamptz,
    successor_reference text, successor_sha256 bytea,
    successor_dedupe_key text, successor_priority integer
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF p_scanned_count < 0 OR p_affected_count < 0 OR
       p_affected_count > p_scanned_count OR
       (p_scanned_count > 0 AND p_cursor_id IS NULL) OR
       pg_catalog.octet_length(p_result_sha256) IS DISTINCT FROM 32 OR
       NOT iam_maintenance_claim_assert_v1(
           p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
           'iam.directory.repair', p_payload_revision,
           p_payload_reference, p_payload_sha256
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM directory repair completion is invalid';
    END IF;
    UPDATE iam_maintenance_checkpoints
    SET cursor_id = p_cursor_id, scanned_count = p_scanned_count,
        affected_count = p_affected_count, error_count = 0,
        result_sha256 = p_result_sha256, last_job_id = p_job_id,
        revision = revision + 1, updated_at = v_now
    WHERE kind = 'iam.directory.repair';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM directory repair checkpoint is missing';
    END IF;
    RETURN QUERY SELECT * FROM iam_maintenance_complete_recurring_v1(
        p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
        'iam.directory.repair', p_payload_revision,
        p_payload_reference, p_payload_sha256,
        p_scanned_count, p_affected_count, 0, p_result_sha256,
        p_successor_job_id
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_directory_repair_finish_v1(
    name,uuid,uuid,uuid,integer,text,bytea,uuid,bigint,bigint,bytea,uuid
) FROM PUBLIC;
