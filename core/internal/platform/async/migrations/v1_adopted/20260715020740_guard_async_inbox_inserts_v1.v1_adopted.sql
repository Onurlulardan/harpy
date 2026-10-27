-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_validate_inbox_insert_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS DISTINCT FROM 'INSERT' OR
       TG_TABLE_SCHEMA IS DISTINCT FROM 'platform' OR
       TG_TABLE_NAME IS DISTINCT FROM 'inbox' OR
       SESSION_USER OPERATOR(pg_catalog.<>) 'harpy_v2_runtime' OR
       NEW.consumer_writer_origin IS DISTINCT FROM 'v2' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5606',
            MESSAGE = 'async inbox consumer authority is invalid';
    END IF;

    IF platform.platform_async_assert_current_authority_v1(
        NEW.scope_kind, NEW.scope_id,
        NEW.consumer_group_id, NEW.consumer_group_generation,
        NEW.consumer_domain, NEW.consumer_generation
    ) IS NULL THEN
        NULL;
    END IF;

    IF NEW.record_kind OPERATOR(pg_catalog.=) 'event_receipt' AND
       NEW.event_id IS NOT NULL AND NEW.event_type IS NOT NULL AND
       NEW.event_version IS NOT NULL AND NEW.event_sequence IS NOT NULL AND
       NEW.through_sequence IS NULL AND NEW.watermark_kind IS NULL AND
       NEW.causation_id IS NULL AND NEW.state_sha256 IS NULL THEN
        IF NOT EXISTS (
            SELECT 1 FROM platform.outbox AS event_row
            WHERE event_row.event_id OPERATOR(pg_catalog.=) NEW.event_id
              AND event_row.tenant_id IS NOT DISTINCT FROM NEW.tenant_id
              AND event_row.scope_kind OPERATOR(pg_catalog.=) NEW.scope_kind
              AND event_row.scope_id OPERATOR(pg_catalog.=) NEW.scope_id
              AND event_row.authority_group_id OPERATOR(pg_catalog.=)
                  NEW.producer_group_id
              AND event_row.authority_group_generation OPERATOR(pg_catalog.=)
                  NEW.producer_group_generation
              AND event_row.authority_domain OPERATOR(pg_catalog.=)
                  NEW.producer_domain
              AND event_row.cutover_generation OPERATOR(pg_catalog.=)
                  NEW.producer_generation
              AND event_row.writer_origin OPERATOR(pg_catalog.=) NEW.writer_origin
              AND event_row.aggregate_type OPERATOR(pg_catalog.=) NEW.aggregate_type
              AND event_row.aggregate_id OPERATOR(pg_catalog.=) NEW.aggregate_id
              AND event_row.event_type OPERATOR(pg_catalog.=) NEW.event_type
              AND event_row.event_version OPERATOR(pg_catalog.=) NEW.event_version
              AND event_row.event_sequence OPERATOR(pg_catalog.=) NEW.event_sequence
            FOR SHARE
        ) THEN
            RAISE EXCEPTION USING ERRCODE = 'P5609',
                MESSAGE = 'async event receipt producer is invalid';
        END IF;
        RETURN NEW;
    END IF;

    IF NEW.record_kind OPERATOR(pg_catalog.=) 'watermark' AND
       NEW.watermark_kind OPERATOR(pg_catalog.=) 'event' AND
       NEW.event_id IS NULL AND NEW.event_type IS NULL AND
       NEW.event_version IS NULL AND NEW.event_sequence IS NULL AND
       NEW.through_sequence IS NOT NULL AND NEW.causation_id IS NOT NULL AND
       NEW.state_sha256 IS NULL THEN
        IF NOT EXISTS (
            SELECT 1
            FROM platform.outbox AS event_row
            JOIN platform.inbox AS receipt
              ON receipt.record_kind OPERATOR(pg_catalog.=) 'event_receipt'
             AND receipt.consumer_name OPERATOR(pg_catalog.=) NEW.consumer_name
             AND receipt.consumer_version OPERATOR(pg_catalog.=) NEW.consumer_version
             AND receipt.event_id OPERATOR(pg_catalog.=) event_row.event_id
             AND receipt.event_type OPERATOR(pg_catalog.=) event_row.event_type
             AND receipt.event_version OPERATOR(pg_catalog.=) event_row.event_version
             AND receipt.event_sequence OPERATOR(pg_catalog.=) event_row.event_sequence
             AND receipt.tenant_id IS NOT DISTINCT FROM NEW.tenant_id
             AND receipt.scope_kind OPERATOR(pg_catalog.=) NEW.scope_kind
             AND receipt.scope_id OPERATOR(pg_catalog.=) NEW.scope_id
             AND receipt.producer_group_id OPERATOR(pg_catalog.=)
                 NEW.producer_group_id
             AND receipt.producer_group_generation OPERATOR(pg_catalog.=)
                 NEW.producer_group_generation
             AND receipt.producer_domain OPERATOR(pg_catalog.=) NEW.producer_domain
             AND receipt.producer_generation OPERATOR(pg_catalog.=)
                 NEW.producer_generation
             AND receipt.writer_origin OPERATOR(pg_catalog.=) NEW.writer_origin
             AND receipt.consumer_group_id OPERATOR(pg_catalog.=)
                 NEW.consumer_group_id
             AND receipt.consumer_group_generation OPERATOR(pg_catalog.=)
                 NEW.consumer_group_generation
             AND receipt.consumer_domain OPERATOR(pg_catalog.=) NEW.consumer_domain
             AND receipt.consumer_generation OPERATOR(pg_catalog.=)
                 NEW.consumer_generation
             AND receipt.aggregate_type OPERATOR(pg_catalog.=) NEW.aggregate_type
             AND receipt.aggregate_id OPERATOR(pg_catalog.=) NEW.aggregate_id
            WHERE event_row.event_id OPERATOR(pg_catalog.=) NEW.causation_id
              AND event_row.tenant_id IS NOT DISTINCT FROM NEW.tenant_id
              AND event_row.scope_kind OPERATOR(pg_catalog.=) NEW.scope_kind
              AND event_row.scope_id OPERATOR(pg_catalog.=) NEW.scope_id
              AND event_row.authority_group_id OPERATOR(pg_catalog.=)
                  NEW.producer_group_id
              AND event_row.authority_group_generation OPERATOR(pg_catalog.=)
                  NEW.producer_group_generation
              AND event_row.authority_domain OPERATOR(pg_catalog.=)
                  NEW.producer_domain
              AND event_row.cutover_generation OPERATOR(pg_catalog.=)
                  NEW.producer_generation
              AND event_row.writer_origin OPERATOR(pg_catalog.=) NEW.writer_origin
              AND event_row.aggregate_type OPERATOR(pg_catalog.=) NEW.aggregate_type
              AND event_row.aggregate_id OPERATOR(pg_catalog.=) NEW.aggregate_id
              AND event_row.event_sequence OPERATOR(pg_catalog.=) NEW.through_sequence
            FOR SHARE OF event_row, receipt
        ) THEN
            RAISE EXCEPTION USING ERRCODE = 'P5611',
                MESSAGE = 'async event watermark producer is invalid';
        END IF;
        RETURN NEW;
    END IF;

    IF NEW.record_kind OPERATOR(pg_catalog.=) 'watermark' AND
       NEW.watermark_kind OPERATOR(pg_catalog.=) 'bootstrap' AND
       NEW.event_id IS NULL AND NEW.event_type IS NULL AND
       NEW.event_version IS NULL AND NEW.event_sequence IS NULL AND
       NEW.through_sequence IS NOT NULL AND NEW.causation_id IS NOT NULL AND
       NEW.state_sha256 IS NOT NULL THEN
        IF NEW.writer_origin IS DISTINCT FROM 'v2' THEN
            RAISE EXCEPTION USING ERRCODE = 'P5607',
                MESSAGE = 'async bootstrap producer origin is invalid';
        END IF;
        IF platform.platform_async_assert_current_authority_v1(
            NEW.scope_kind, NEW.scope_id,
            NEW.producer_group_id, NEW.producer_group_generation,
            NEW.producer_domain, NEW.producer_generation
        ) IS NULL THEN
            NULL;
        END IF;
        RETURN NEW;
    END IF;

    RAISE EXCEPTION USING ERRCODE = 'P5608',
        MESSAGE = 'async inbox record shape is invalid';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_validate_inbox_insert_v1()
FROM PUBLIC;

CREATE TRIGGER inbox_authority_insert_guard
BEFORE INSERT ON platform.inbox
FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_validate_inbox_insert_v1();
ALTER TABLE platform.inbox ENABLE ALWAYS TRIGGER inbox_authority_insert_guard;
