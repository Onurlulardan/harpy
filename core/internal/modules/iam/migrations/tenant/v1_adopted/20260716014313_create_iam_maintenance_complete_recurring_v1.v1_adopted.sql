-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_complete_recurring_v1(
    p_scope_schema name, p_tenant_id uuid, p_job_id uuid,
    p_attempt_id uuid, p_kind text, p_payload_revision integer,
    p_payload_reference text, p_payload_sha256 bytea,
    p_scanned_count bigint, p_affected_count bigint, p_error_count bigint,
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
    v_scheduled_run_at timestamptz;
    v_cadence_seconds integer;
    v_next_run timestamptz;
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_outcome text;
BEGIN
    v_outcome := p_kind || '.completed';
    IF p_scanned_count < 0 OR p_affected_count < 0 OR p_error_count < 0 OR
       p_affected_count > p_scanned_count OR
       pg_catalog.octet_length(p_result_sha256) IS DISTINCT FROM 32 OR
       p_successor_job_id IS NULL OR
       pg_catalog.substring(p_successor_job_id::text, 15, 1) <> '7' OR
       p_kind = 'iam.profile.backfill' OR NOT iam_maintenance_claim_assert_v1(
           p_scope_schema, p_tenant_id, p_job_id, p_attempt_id, p_kind,
           p_payload_revision, p_payload_reference, p_payload_sha256
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM recurring maintenance completion is invalid';
    END IF;
    SELECT schedule.scheduled_run_at, schedule.cadence_seconds
    INTO v_scheduled_run_at, v_cadence_seconds
    FROM iam_maintenance_schedules AS schedule
    WHERE schedule.kind = p_kind AND schedule.current_job_id = p_job_id
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM maintenance schedule claim changed';
    END IF;
    v_next_run := v_scheduled_run_at +
        v_cadence_seconds * INTERVAL '1 second';
    UPDATE iam_maintenance_schedules
    SET current_job_id = p_successor_job_id, scheduled_run_at = v_next_run,
        revision = revision + 1, updated_at = v_now
    WHERE kind = p_kind AND current_job_id = p_job_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM maintenance schedule advance lost';
    END IF;
    INSERT INTO iam_maintenance_receipts (
        job_id, attempt_id, kind, payload_revision, payload_reference,
        payload_sha256, outcome_code, scanned_count, affected_count,
        error_count, result_sha256, successor_job_id, successor_run_at,
        successor_payload_reference, successor_payload_sha256,
        successor_dedupe_key, successor_priority, created_at
    ) VALUES (
        p_job_id, p_attempt_id, p_kind, p_payload_revision,
        p_payload_reference, p_payload_sha256, v_outcome, p_scanned_count,
        p_affected_count, p_error_count, p_result_sha256,
        p_successor_job_id, v_next_run, p_payload_reference, p_payload_sha256,
        p_kind || ':' || p_successor_job_id::text,
        0, v_now
    );
    INSERT INTO iam_maintenance_audit (
        job_id, attempt_id, kind, outcome_code, affected_count,
        error_count, result_sha256, occurred_at
    ) VALUES (
        p_job_id, p_attempt_id, p_kind, v_outcome, p_affected_count,
        p_error_count, p_result_sha256, v_now
    );
    INSERT INTO iam_maintenance_metrics (
        kind, completed_runs, affected_total, error_total, last_job_id,
        last_result_sha256, last_completed_at, revision
    ) VALUES (
        p_kind, 1, p_affected_count, p_error_count, p_job_id,
        p_result_sha256, v_now, 1
    ) ON CONFLICT (kind) DO UPDATE SET
        completed_runs = iam_maintenance_metrics.completed_runs + 1,
        affected_total = iam_maintenance_metrics.affected_total +
            EXCLUDED.affected_total,
        error_total = iam_maintenance_metrics.error_total + EXCLUDED.error_total,
        last_job_id = EXCLUDED.last_job_id,
        last_result_sha256 = EXCLUDED.last_result_sha256,
        last_completed_at = EXCLUDED.last_completed_at,
        revision = iam_maintenance_metrics.revision + 1;
    RETURN QUERY SELECT
        p_successor_job_id, v_next_run, p_payload_reference, p_payload_sha256,
        p_kind || ':' || p_successor_job_id::text,
        0;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_complete_recurring_v1(
    name,uuid,uuid,uuid,text,integer,text,bytea,bigint,bigint,bigint,bytea,uuid
) FROM PUBLIC;
