-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_guard_inbox_stream_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_latest bigint;
BEGIN
    IF TG_OP IS DISTINCT FROM 'INSERT' OR
       TG_TABLE_SCHEMA IS DISTINCT FROM 'platform' OR
       TG_TABLE_NAME IS DISTINCT FROM 'inbox' OR
       SESSION_USER OPERATOR(pg_catalog.<>) 'forgecrud_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5630',
            MESSAGE = 'async inbox stream authority is invalid';
    END IF;

    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        pg_catalog.concat_ws(E'\x1f',
            'platform_async_inbox_v1', NEW.consumer_name,
            NEW.consumer_version::text, NEW.scope_kind, NEW.scope_id::text,
            NEW.aggregate_type, NEW.aggregate_id::text
        ), 0
    ));

    IF NEW.record_kind OPERATOR(pg_catalog.=) 'event_receipt' THEN
        IF EXISTS (
            SELECT 1 FROM platform.inbox AS duplicate
            WHERE duplicate.consumer_name OPERATOR(pg_catalog.=) NEW.consumer_name
              AND duplicate.consumer_version OPERATOR(pg_catalog.=)
                  NEW.consumer_version
              AND duplicate.event_id OPERATOR(pg_catalog.=) NEW.event_id
        ) THEN
            RETURN NEW;
        END IF;

        SELECT pg_catalog.max(watermark.through_sequence)
        INTO v_latest
        FROM platform.inbox AS watermark
        WHERE watermark.record_kind OPERATOR(pg_catalog.=) 'watermark'
          AND watermark.consumer_name OPERATOR(pg_catalog.=) NEW.consumer_name
          AND watermark.consumer_version OPERATOR(pg_catalog.=) NEW.consumer_version
          AND watermark.scope_kind OPERATOR(pg_catalog.=) NEW.scope_kind
          AND watermark.scope_id OPERATOR(pg_catalog.=) NEW.scope_id
          AND watermark.aggregate_type OPERATOR(pg_catalog.=) NEW.aggregate_type
          AND watermark.aggregate_id OPERATOR(pg_catalog.=) NEW.aggregate_id;

        IF NEW.event_sequence IS DISTINCT FROM
           (COALESCE(v_latest, 0::bigint) OPERATOR(pg_catalog.+) 1) THEN
            RAISE EXCEPTION USING ERRCODE = 'P5631',
                MESSAGE = 'async inbox event sequence is not contiguous';
        END IF;
        RETURN NEW;
    END IF;

    IF NEW.record_kind OPERATOR(pg_catalog.=) 'watermark' AND
       NEW.watermark_kind OPERATOR(pg_catalog.=) 'event' THEN
        IF EXISTS (
            SELECT 1 FROM platform.inbox AS duplicate
            WHERE duplicate.record_kind OPERATOR(pg_catalog.=) 'watermark'
              AND duplicate.consumer_name OPERATOR(pg_catalog.=) NEW.consumer_name
              AND duplicate.consumer_version OPERATOR(pg_catalog.=)
                  NEW.consumer_version
              AND duplicate.scope_kind OPERATOR(pg_catalog.=) NEW.scope_kind
              AND duplicate.scope_id OPERATOR(pg_catalog.=) NEW.scope_id
              AND duplicate.aggregate_type OPERATOR(pg_catalog.=) NEW.aggregate_type
              AND duplicate.aggregate_id OPERATOR(pg_catalog.=) NEW.aggregate_id
              AND duplicate.through_sequence OPERATOR(pg_catalog.=)
                  NEW.through_sequence
        ) THEN
            RETURN NEW;
        END IF;

        SELECT pg_catalog.max(watermark.through_sequence)
        INTO v_latest
        FROM platform.inbox AS watermark
        WHERE watermark.record_kind OPERATOR(pg_catalog.=) 'watermark'
          AND watermark.consumer_name OPERATOR(pg_catalog.=) NEW.consumer_name
          AND watermark.consumer_version OPERATOR(pg_catalog.=) NEW.consumer_version
          AND watermark.scope_kind OPERATOR(pg_catalog.=) NEW.scope_kind
          AND watermark.scope_id OPERATOR(pg_catalog.=) NEW.scope_id
          AND watermark.aggregate_type OPERATOR(pg_catalog.=) NEW.aggregate_type
          AND watermark.aggregate_id OPERATOR(pg_catalog.=) NEW.aggregate_id;

        IF NEW.through_sequence IS DISTINCT FROM
           (COALESCE(v_latest, 0::bigint) OPERATOR(pg_catalog.+) 1) THEN
            RAISE EXCEPTION USING ERRCODE = 'P5632',
                MESSAGE = 'async inbox event watermark is not contiguous';
        END IF;
        RETURN NEW;
    END IF;

    IF NEW.record_kind OPERATOR(pg_catalog.=) 'watermark' AND
       NEW.watermark_kind OPERATOR(pg_catalog.=) 'bootstrap' THEN
        IF EXISTS (
            SELECT 1 FROM platform.inbox AS stream_row
            WHERE stream_row.consumer_name OPERATOR(pg_catalog.=) NEW.consumer_name
              AND stream_row.consumer_version OPERATOR(pg_catalog.=)
                  NEW.consumer_version
              AND stream_row.scope_kind OPERATOR(pg_catalog.=) NEW.scope_kind
              AND stream_row.scope_id OPERATOR(pg_catalog.=) NEW.scope_id
              AND stream_row.aggregate_type OPERATOR(pg_catalog.=)
                  NEW.aggregate_type
              AND stream_row.aggregate_id OPERATOR(pg_catalog.=) NEW.aggregate_id
        ) THEN
            RAISE EXCEPTION USING ERRCODE = 'P5633',
                MESSAGE = 'async inbox bootstrap requires an empty stream';
        END IF;
        RETURN NEW;
    END IF;

    RAISE EXCEPTION USING ERRCODE = 'P5634',
        MESSAGE = 'async inbox stream record is invalid';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_guard_inbox_stream_v1()
FROM PUBLIC;

CREATE TRIGGER inbox_stream_insert_guard
BEFORE INSERT ON platform.inbox
FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_guard_inbox_stream_v1();
ALTER TABLE platform.inbox ENABLE ALWAYS TRIGGER inbox_stream_insert_guard;
