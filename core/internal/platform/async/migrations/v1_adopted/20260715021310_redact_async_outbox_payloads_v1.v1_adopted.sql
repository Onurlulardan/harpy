-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_redact_outbox_payloads_v1(p_batch_size integer)
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
        SELECT event.event_id
        FROM platform.outbox AS event
        WHERE event.delivery_state = 'delivered'
          AND event.payload_canonical IS NOT NULL
          AND event.payload_redacted_at IS NULL
          AND event.delivered_at <= v_now - interval '30 days'
        ORDER BY event.delivered_at, event.event_id
        FOR UPDATE SKIP LOCKED
        LIMIT p_batch_size
    )
    UPDATE platform.outbox AS event
    SET payload_canonical = NULL,
        payload_redacted_at = v_now,
        updated_at = v_now
    FROM candidates
    WHERE event.event_id = candidates.event_id;
    GET DIAGNOSTICS v_count = ROW_COUNT;
    RETURN v_count;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_redact_outbox_payloads_v1(integer) FROM PUBLIC;
