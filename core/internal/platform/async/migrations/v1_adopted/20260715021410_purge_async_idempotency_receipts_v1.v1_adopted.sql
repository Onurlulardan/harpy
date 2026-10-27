-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_purge_idempotency_receipts_v1(
    p_batch_size integer
)
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
            MESSAGE = 'async purge batch size must be between 1 and 500';
    END IF;

    WITH candidates AS (
        SELECT receipt.receipt_id
        FROM platform.idempotency_receipts AS receipt
        WHERE receipt.retain_until <= v_now
          AND (
              receipt.retention_class <> 'async_handle' OR EXISTS (
                  SELECT 1
                  FROM platform.jobs AS job
                  WHERE job.scope_kind = receipt.scope_kind
                    AND job.scope_id = receipt.scope_id
                    AND job.id = receipt.async_job_id
                    AND job.state IN ('succeeded', 'dead', 'cancelled')
                    AND job.completed_at <= v_now - interval '24 hours'
              )
          )
          AND NOT EXISTS (
              SELECT 1
              FROM platform.job_attempts AS attempt
              WHERE attempt.effect_receipt_id = receipt.receipt_id
          )
        ORDER BY receipt.retain_until, receipt.receipt_id
        FOR UPDATE SKIP LOCKED
        LIMIT p_batch_size
    )
    DELETE FROM platform.idempotency_receipts AS receipt
    USING candidates
    WHERE receipt.receipt_id = candidates.receipt_id;
    GET DIAGNOSTICS v_count = ROW_COUNT;
    RETURN v_count;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_purge_idempotency_receipts_v1(integer) FROM PUBLIC;
