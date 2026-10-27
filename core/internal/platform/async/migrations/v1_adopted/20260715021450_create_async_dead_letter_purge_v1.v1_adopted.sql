-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_purge_dead_letters_v1(p_batch_size integer)
RETURNS integer
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_now timestamptz;
    v_root_id uuid;
    v_action_id uuid;
    v_deleted integer;
    v_count integer;
BEGIN
    v_now := pg_catalog.clock_timestamp();
    v_count := 0;
    IF p_batch_size IS NULL OR p_batch_size NOT BETWEEN 1 AND 500 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'async purge batch size must be between 1 and 500';
    END IF;

    FOR v_root_id IN
        SELECT dead_letter.id
        FROM platform.dead_letters AS dead_letter
        WHERE dead_letter.record_kind = 'root'
          AND dead_letter.metadata_expires_at <= v_now
          AND dead_letter.payload_canonical IS NULL
          AND dead_letter.payload_redacted_at IS NOT NULL
        ORDER BY dead_letter.metadata_expires_at, dead_letter.id
        FOR UPDATE SKIP LOCKED
        LIMIT p_batch_size
    LOOP
        FOR v_action_id IN
            SELECT action.id
            FROM platform.dead_letters AS action
            WHERE action.root_id = v_root_id
              AND action.id <> v_root_id
            ORDER BY action.entry_sequence DESC
            FOR UPDATE
        LOOP
            DELETE FROM platform.dead_letters AS action
            WHERE action.id = v_action_id;
        END LOOP;

        DELETE FROM platform.dead_letters AS root
        WHERE root.id = v_root_id
          AND root.root_id = root.id;
        GET DIAGNOSTICS v_deleted = ROW_COUNT;
        v_count := v_count + v_deleted;
    END LOOP;

    RETURN v_count;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_purge_dead_letters_v1(integer)
    FROM PUBLIC;
