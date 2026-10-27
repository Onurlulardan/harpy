-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_purge_inbox_v1(p_batch_size integer)
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
        SELECT inbox_row.id
        FROM platform.inbox AS inbox_row
        WHERE inbox_row.recorded_at <= v_now - interval '180 days'
          AND (
              inbox_row.record_kind = 'event_receipt' OR
              (inbox_row.record_kind = 'watermark' AND EXISTS (
                  SELECT 1
                  FROM platform.inbox AS newer
                  WHERE newer.record_kind = 'watermark'
                    AND newer.consumer_name = inbox_row.consumer_name
                    AND newer.consumer_version = inbox_row.consumer_version
                    AND newer.scope_kind = inbox_row.scope_kind
                    AND newer.scope_id = inbox_row.scope_id
                    AND newer.aggregate_type = inbox_row.aggregate_type
                    AND newer.aggregate_id = inbox_row.aggregate_id
                    AND newer.through_sequence > inbox_row.through_sequence
              ))
          )
        ORDER BY inbox_row.recorded_at, inbox_row.id
        FOR UPDATE SKIP LOCKED
        LIMIT p_batch_size
    )
    DELETE FROM platform.inbox AS inbox_row
    USING candidates
    WHERE inbox_row.id = candidates.id;
    GET DIAGNOSTICS v_count = ROW_COUNT;
    RETURN v_count;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_purge_inbox_v1(integer)
    FROM PUBLIC;
