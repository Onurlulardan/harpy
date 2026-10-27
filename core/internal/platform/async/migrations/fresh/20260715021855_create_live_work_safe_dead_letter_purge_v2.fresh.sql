-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_purge_dead_letters_v2(p_batch_size integer)
RETURNS integer
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_now timestamptz;
    v_root_id uuid;
    v_root platform.dead_letters%ROWTYPE;
    v_count integer;
    v_changed integer;
BEGIN
    v_now := pg_catalog.clock_timestamp();
    v_count := 0;
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'async retention session authority is invalid';
    END IF;
    IF p_batch_size IS NULL OR p_batch_size NOT BETWEEN 1 AND 500 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'async purge batch size must be between 1 and 500';
    END IF;

    FOR v_root_id IN
        SELECT root.id
        FROM platform.dead_letters AS root
        WHERE root.record_kind = 'root'
          AND root.metadata_expires_at <= v_now
          AND root.payload_canonical IS NULL
          AND root.payload_redacted_at IS NOT NULL
          AND NOT platform.platform_async_has_live_event_gap_root_v1(root)
          AND NOT platform.platform_async_has_live_stale_job_root_v1(root)
        ORDER BY root.metadata_expires_at,root.id
        LIMIT p_batch_size
    LOOP
        PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
            'forgecrud:deadletter:' || v_root_id::text,0));
        SELECT root.* INTO v_root
        FROM platform.dead_letters AS root
        WHERE root.id = v_root_id AND root.record_kind = 'root'
          AND root.metadata_expires_at <= v_now
          AND root.payload_canonical IS NULL
          AND root.payload_redacted_at IS NOT NULL
        FOR UPDATE;
        IF NOT FOUND OR
           platform.platform_async_has_live_event_gap_root_v1(v_root) OR
           platform.platform_async_has_live_stale_job_root_v1(v_root) THEN
            CONTINUE;
        END IF;
        DELETE FROM platform.dead_letters AS action
        WHERE action.root_id = v_root_id AND action.id <> v_root_id;
        DELETE FROM platform.dead_letters AS root
        WHERE root.id = v_root_id AND root.root_id = root.id;
        GET DIAGNOSTICS v_changed = ROW_COUNT;
        v_count := v_count + v_changed;
    END LOOP;
    RETURN v_count;
END;
$function$;

DROP FUNCTION platform.platform_async_purge_dead_letters_v1(integer);
CREATE FUNCTION platform.platform_async_purge_dead_letters_v1(p_batch_size integer)
RETURNS integer
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE = '55000',
        MESSAGE = 'unsafe dead-letter purge v1 is retired; use v2';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_purge_dead_letters_v2(integer) FROM PUBLIC;
REVOKE ALL ON FUNCTION
    platform.platform_async_purge_dead_letters_v1(integer) FROM PUBLIC;
