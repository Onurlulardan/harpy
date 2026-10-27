-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_redact_job_payloads_v1(p_batch_size integer)
RETURNS integer
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_now timestamptz;
    v_count integer;
BEGIN
    v_now := pg_catalog.clock_timestamp();
    IF p_batch_size IS NULL OR p_batch_size NOT BETWEEN 1 AND 500 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'async redaction batch size must be between 1 and 500';
    END IF;

    WITH candidates AS (
        SELECT job.id
        FROM platform.jobs AS job
        WHERE job.state IN ('succeeded', 'dead', 'cancelled')
          AND job.payload_ref IS NOT NULL
          AND job.payload_redacted_at IS NULL
          AND job.completed_at <= v_now - interval '30 days'
        ORDER BY job.completed_at, job.id
        FOR UPDATE SKIP LOCKED
        LIMIT p_batch_size
    )
    UPDATE platform.jobs AS job
    SET payload_ref = NULL,
        payload_redacted_at = v_now,
        updated_at = v_now
    FROM candidates
    WHERE job.id = candidates.id;
    GET DIAGNOSTICS v_count = ROW_COUNT;
    RETURN v_count;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_redact_job_payloads_v1(integer) FROM PUBLIC;
