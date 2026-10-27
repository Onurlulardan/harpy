-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_profile_backfill_receipt_v1(
    p_job_id uuid, p_attempt_id uuid, p_payload_revision integer,
    p_payload_reference text, p_payload_sha256 bytea,
    p_scanned_count bigint, p_affected_count bigint, p_error_count bigint,
    p_result_sha256 bytea, p_successor_job_id uuid,
    p_successor_run_at timestamptz, p_successor_reference text,
    p_successor_sha256 bytea, p_successor_dedupe_key text
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
    v_outcome text;
BEGIN
    v_outcome := 'iam.profile.backfill.completed';
    IF p_scanned_count < 0 OR p_affected_count < 0 OR p_error_count < 0 OR
       p_affected_count > p_scanned_count OR
       pg_catalog.octet_length(p_result_sha256) IS DISTINCT FROM 32 OR
       ((p_successor_job_id IS NULL) <> (p_successor_run_at IS NULL)) OR
       ((p_successor_job_id IS NULL) <> (p_successor_reference IS NULL)) OR
       ((p_successor_job_id IS NULL) <> (p_successor_sha256 IS NULL)) OR
       ((p_successor_job_id IS NULL) <> (p_successor_dedupe_key IS NULL)) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM profile backfill receipt is invalid';
    END IF;
    INSERT INTO iam_maintenance_receipts (
        job_id, attempt_id, kind, payload_revision, payload_reference,
        payload_sha256, outcome_code, scanned_count, affected_count,
        error_count, result_sha256, successor_job_id, successor_run_at,
        successor_payload_reference, successor_payload_sha256,
        successor_dedupe_key, successor_priority, created_at
    ) VALUES (
        p_job_id, p_attempt_id, 'iam.profile.backfill', p_payload_revision,
        p_payload_reference, p_payload_sha256, v_outcome, p_scanned_count,
        p_affected_count, p_error_count, p_result_sha256,
        p_successor_job_id, p_successor_run_at, p_successor_reference,
        p_successor_sha256, p_successor_dedupe_key,
        CASE WHEN p_successor_job_id IS NULL THEN NULL ELSE 0 END, v_now
    );
    INSERT INTO iam_maintenance_audit (
        job_id, attempt_id, kind, outcome_code, affected_count,
        error_count, result_sha256, occurred_at
    ) VALUES (
        p_job_id, p_attempt_id, 'iam.profile.backfill', v_outcome,
        p_affected_count, p_error_count, p_result_sha256, v_now
    );
    INSERT INTO iam_maintenance_metrics (
        kind, completed_runs, affected_total, error_total, last_job_id,
        last_result_sha256, last_completed_at, revision
    ) VALUES (
        'iam.profile.backfill', 1, p_affected_count, p_error_count,
        p_job_id, p_result_sha256, v_now, 1
    ) ON CONFLICT (kind) DO UPDATE SET
        completed_runs = iam_maintenance_metrics.completed_runs + 1,
        affected_total = iam_maintenance_metrics.affected_total +
            EXCLUDED.affected_total,
        error_total = iam_maintenance_metrics.error_total + EXCLUDED.error_total,
        last_job_id = EXCLUDED.last_job_id,
        last_result_sha256 = EXCLUDED.last_result_sha256,
        last_completed_at = EXCLUDED.last_completed_at,
        revision = iam_maintenance_metrics.revision + 1;
    RETURN QUERY SELECT p_successor_job_id, p_successor_run_at,
        p_successor_reference, p_successor_sha256,
        p_successor_dedupe_key,
        CASE WHEN p_successor_job_id IS NULL THEN NULL ELSE 0 END;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_profile_backfill_receipt_v1(
    uuid,uuid,integer,text,bytea,bigint,bigint,bigint,bytea,
    uuid,timestamptz,text,bytea,text
) FROM PUBLIC;
