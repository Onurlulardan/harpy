-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_directory_checksum_finish_v1(
    p_scope_schema name, p_tenant_id uuid, p_job_id uuid,
    p_attempt_id uuid, p_payload_revision integer,
    p_payload_reference text, p_payload_sha256 bytea,
    p_scanned_count bigint, p_drift_count bigint, p_result_sha256 bytea,
    p_key_evidence jsonb, p_successor_job_id uuid
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
    v_entry jsonb;
    v_key_count bigint;
BEGIN
    IF p_key_evidence IS NULL OR
       pg_catalog.jsonb_typeof(p_key_evidence) IS DISTINCT FROM 'array' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM directory checksum key evidence is invalid';
    END IF;
    IF p_scanned_count < 0 OR p_drift_count < 0 OR
       p_drift_count > p_scanned_count OR
       pg_catalog.octet_length(p_result_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.jsonb_array_length(p_key_evidence) NOT BETWEEN 1 AND 65 OR
       NOT iam_maintenance_claim_assert_v1(
           p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
           'iam.directory.checksum', p_payload_revision,
           p_payload_reference, p_payload_sha256
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM directory checksum completion is invalid';
    END IF;
    IF EXISTS (
        SELECT 1
        FROM pg_catalog.jsonb_array_elements(p_key_evidence) AS item(value)
        WHERE pg_catalog.jsonb_typeof(item.value) <> 'object'
           OR item.value <> pg_catalog.jsonb_build_object(
                'key_id', item.value ->> 'key_id',
                'count', item.value -> 'count'
              )
           OR NOT (
                (pg_catalog.octet_length(item.value ->> 'key_id')
                    BETWEEN 1 AND 128 AND
                 item.value ->> 'key_id' ~ '^[a-z0-9][a-z0-9._-]{0,127}$') OR
                iam_hmac_key_id_valid_v1(item.value ->> 'key_id')
              )
           OR pg_catalog.jsonb_typeof(item.value -> 'count')
                IS DISTINCT FROM 'number'
           OR item.value ->> 'count' !~ '^(0|[1-9][0-9]*)$'
           OR (item.value ->> 'count')::numeric > 9223372036854775807
    ) OR (
        SELECT pg_catalog.count(DISTINCT item.value ->> 'key_id')
        FROM pg_catalog.jsonb_array_elements(p_key_evidence) AS item(value)
    ) <> pg_catalog.jsonb_array_length(p_key_evidence) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM directory checksum key evidence is invalid';
    END IF;
    FOR v_entry IN
        SELECT item.value
        FROM pg_catalog.jsonb_array_elements(p_key_evidence) AS item(value)
        ORDER BY item.value ->> 'key_id'
    LOOP
        v_key_count := (v_entry ->> 'count')::bigint;
        INSERT INTO iam_directory_key_evidence (
            lookup_key_id, last_job_id, projected_row_count, zero_rows_since,
            last_result_sha256, revision, updated_at
        ) VALUES (
            v_entry ->> 'key_id', p_job_id, v_key_count,
            CASE WHEN v_key_count = 0 THEN v_now ELSE NULL END,
            p_result_sha256, 1, v_now
        ) ON CONFLICT (lookup_key_id) DO UPDATE SET
            last_job_id = EXCLUDED.last_job_id,
            projected_row_count = EXCLUDED.projected_row_count,
            zero_rows_since = CASE
                WHEN EXCLUDED.projected_row_count > 0 THEN NULL
                ELSE COALESCE(iam_directory_key_evidence.zero_rows_since, v_now)
            END,
            last_result_sha256 = EXCLUDED.last_result_sha256,
            revision = iam_directory_key_evidence.revision + 1,
            updated_at = v_now;
    END LOOP;
    UPDATE iam_maintenance_checkpoints
    SET cursor_id = NULL, scanned_count = p_scanned_count,
        affected_count = p_drift_count, error_count = 0,
        result_sha256 = p_result_sha256, last_job_id = p_job_id,
        revision = revision + 1, updated_at = v_now
    WHERE kind = 'iam.directory.checksum';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM directory checksum checkpoint is missing';
    END IF;
    RETURN QUERY SELECT * FROM iam_maintenance_complete_recurring_v1(
        p_scope_schema, p_tenant_id, p_job_id, p_attempt_id,
        'iam.directory.checksum', p_payload_revision,
        p_payload_reference, p_payload_sha256,
        p_scanned_count, p_drift_count, 0, p_result_sha256,
        p_successor_job_id
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_directory_checksum_finish_v1(
    name,uuid,uuid,uuid,integer,text,bytea,bigint,bigint,bytea,jsonb,uuid
) FROM PUBLIC;
