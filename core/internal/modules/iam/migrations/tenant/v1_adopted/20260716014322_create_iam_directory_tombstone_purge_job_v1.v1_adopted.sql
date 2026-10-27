-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_directory_tombstone_purge_v1(
    p_scope_schema name, p_tenant_id uuid, p_job_id uuid,
    p_attempt_id uuid, p_payload_revision integer,
    p_payload_reference text, p_payload_sha256 bytea,
    p_scanned_count bigint, p_purged_count bigint,
    p_successor_job_id uuid
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
    v_result bytea;
BEGIN
    IF p_scanned_count < 0 OR p_purged_count < 0 OR
       p_purged_count > p_scanned_count OR
       NOT iam_maintenance_claim_assert_v1(
           p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
           'iam.directory.tombstone-purge', p_payload_revision,
           p_payload_reference, p_payload_sha256
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM directory tombstone purge completion is invalid';
    END IF;
    v_result := pg_catalog.sha256(pg_catalog.convert_to(
        'iam.directory.tombstone-purge:' || p_scanned_count::text || ':' ||
        p_purged_count::text, 'UTF8'
    ));
    UPDATE iam_maintenance_checkpoints
    SET cursor_id = NULL, scanned_count = p_scanned_count,
        affected_count = p_purged_count, error_count = 0,
        result_sha256 = v_result, last_job_id = p_job_id,
        revision = revision + 1,
        updated_at = pg_catalog.transaction_timestamp()
    WHERE kind = 'iam.directory.tombstone-purge';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM directory purge checkpoint is missing';
    END IF;
    RETURN QUERY SELECT * FROM iam_maintenance_complete_recurring_v1(
        p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
        'iam.directory.tombstone-purge', p_payload_revision,
        p_payload_reference, p_payload_sha256,
        p_scanned_count, p_purged_count, 0, v_result, p_successor_job_id
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_directory_tombstone_purge_v1(
    name,uuid,uuid,uuid,integer,text,bytea,bigint,bigint,uuid
) FROM PUBLIC;
