-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_redact_outbox_payloads_v2(
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
    v_changed integer;
    v_event_id uuid;
    v_event platform.outbox%ROWTYPE;
BEGIN
    v_now := pg_catalog.clock_timestamp();
    v_count := 0;
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'async retention session authority is invalid';
    END IF;
    IF p_batch_size IS NULL OR p_batch_size NOT BETWEEN 1 AND 500 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'async redaction batch size must be between 1 and 500';
    END IF;

    FOR v_event_id IN
        SELECT event.event_id
        FROM platform.outbox AS event
        WHERE event.delivery_state = 'delivered'
          AND event.payload_canonical IS NOT NULL
          AND event.payload_redacted_at IS NULL
          AND event.delivered_at <= v_now - interval '30 days'
          AND NOT platform.platform_async_has_live_event_gap_v1(event)
        ORDER BY event.delivered_at, event.event_id
        LIMIT p_batch_size
    LOOP
        SELECT event.* INTO v_event
        FROM platform.outbox AS event
        WHERE event.event_id = v_event_id
          AND event.delivery_state = 'delivered'
          AND event.payload_canonical IS NOT NULL
          AND event.payload_redacted_at IS NULL
          AND event.delivered_at <= v_now - interval '30 days'
        FOR UPDATE;
        IF NOT FOUND OR platform.platform_async_has_live_event_gap_v1(v_event) THEN
            CONTINUE;
        END IF;
        UPDATE platform.outbox AS event
        SET payload_canonical = NULL,payload_redacted_at = v_now,updated_at = v_now
        WHERE event.event_id = v_event_id;
        GET DIAGNOSTICS v_changed = ROW_COUNT;
        v_count := v_count + v_changed;
    END LOOP;
    RETURN v_count;
END;
$function$;

CREATE FUNCTION platform.platform_async_redact_dead_letter_payloads_v2(
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
    v_changed integer;
    v_root_id uuid;
    v_root platform.dead_letters%ROWTYPE;
BEGIN
    v_now := pg_catalog.clock_timestamp();
    v_count := 0;
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'async retention session authority is invalid';
    END IF;
    IF p_batch_size IS NULL OR p_batch_size NOT BETWEEN 1 AND 500 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'async redaction batch size must be between 1 and 500';
    END IF;

    FOR v_root_id IN
        SELECT root.id
        FROM platform.dead_letters AS root
        WHERE root.record_kind = 'root'
          AND root.payload_canonical IS NOT NULL
          AND root.payload_redacted_at IS NULL
          AND root.full_payload_expires_at <= v_now
          AND NOT platform.platform_async_has_live_event_gap_root_v1(root)
          AND NOT platform.platform_async_has_live_stale_job_root_v1(root)
        ORDER BY root.full_payload_expires_at, root.id
        LIMIT p_batch_size
    LOOP
        PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
            'harpy:deadletter:' || v_root_id::text,0));
        SELECT root.* INTO v_root
        FROM platform.dead_letters AS root
        WHERE root.id = v_root_id
          AND root.record_kind = 'root'
          AND root.payload_canonical IS NOT NULL
          AND root.payload_redacted_at IS NULL
          AND root.full_payload_expires_at <= v_now
        FOR UPDATE;
        IF NOT FOUND OR
           platform.platform_async_has_live_event_gap_root_v1(v_root) OR
           platform.platform_async_has_live_stale_job_root_v1(v_root) THEN
            CONTINUE;
        END IF;
        UPDATE platform.dead_letters AS root
        SET payload_canonical = NULL,payload_redacted_at = v_now
        WHERE root.id = v_root_id;
        GET DIAGNOSTICS v_changed = ROW_COUNT;
        v_count := v_count + v_changed;
    END LOOP;
    RETURN v_count;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_redact_outbox_payloads_v2(integer)
FROM PUBLIC;
REVOKE ALL ON FUNCTION
    platform.platform_async_redact_dead_letter_payloads_v2(integer)
FROM PUBLIC;
