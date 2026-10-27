-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_redact_dead_letter_payloads_v1(
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
            MESSAGE = 'async redaction batch size must be between 1 and 500';
    END IF;

    WITH candidates AS (
        SELECT dead_letter.id
        FROM platform.dead_letters AS dead_letter
        WHERE dead_letter.record_kind = 'root'
          AND dead_letter.payload_canonical IS NOT NULL
          AND dead_letter.payload_redacted_at IS NULL
          AND dead_letter.full_payload_expires_at <= v_now
        ORDER BY dead_letter.full_payload_expires_at, dead_letter.id
        FOR UPDATE SKIP LOCKED
        LIMIT p_batch_size
    )
    UPDATE platform.dead_letters AS dead_letter
    SET payload_canonical = NULL,
        payload_redacted_at = v_now
    FROM candidates
    WHERE dead_letter.id = candidates.id;
    GET DIAGNOSTICS v_count = ROW_COUNT;
    RETURN v_count;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_redact_dead_letter_payloads_v1(integer) FROM PUBLIC;
