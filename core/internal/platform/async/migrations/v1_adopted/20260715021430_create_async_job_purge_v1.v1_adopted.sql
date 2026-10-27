-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_purge_jobs_v1(p_batch_size integer)
RETURNS integer
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_now timestamptz;
    v_job_id uuid;
    v_deleted integer;
    v_count integer;
BEGIN
    v_now := pg_catalog.clock_timestamp();
    v_count := 0;
    IF p_batch_size IS NULL OR p_batch_size NOT BETWEEN 1 AND 500 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'async purge batch size must be between 1 and 500';
    END IF;

    FOR v_job_id IN
        SELECT job.id
        FROM platform.jobs AS job
        WHERE job.state IN ('succeeded', 'dead', 'cancelled')
          AND job.payload_ref IS NULL
          AND job.payload_redacted_at IS NOT NULL
          AND job.retain_until <= v_now
          AND job.retention_release_revision > 0
          AND job.retention_release_sha256 IS NOT NULL
          AND job.retention_released_at IS NOT NULL
          AND job.retention_released_at <= v_now
          AND NOT EXISTS (
              SELECT 1
              FROM platform.job_attempts AS attempt
              WHERE attempt.job_id = job.id
                AND attempt.finished_at IS NULL
          )
          AND NOT EXISTS (
              SELECT 1
              FROM platform.idempotency_receipts AS receipt
              WHERE receipt.async_job_id = job.id
          )
          AND NOT EXISTS (
              SELECT 1
              FROM platform.jobs AS retry
              WHERE retry.retry_of_job_id = job.id
          )
        ORDER BY job.retain_until, job.id
        FOR UPDATE SKIP LOCKED
        LIMIT p_batch_size
    LOOP
        DELETE FROM platform.job_attempts AS attempt
        WHERE attempt.job_id = v_job_id;

        DELETE FROM platform.jobs AS job
        WHERE job.id = v_job_id;
        GET DIAGNOSTICS v_deleted = ROW_COUNT;
        v_count := v_count + v_deleted;
    END LOOP;

    RETURN v_count;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_purge_jobs_v1(integer)
    FROM PUBLIC;
