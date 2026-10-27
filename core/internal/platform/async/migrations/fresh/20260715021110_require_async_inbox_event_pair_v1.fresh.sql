-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_require_inbox_event_pair_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF NEW.record_kind OPERATOR(pg_catalog.=) 'event_receipt' AND NOT EXISTS (
        SELECT 1 FROM platform.inbox AS pair
        WHERE pair.record_kind OPERATOR(pg_catalog.=) 'watermark'
          AND pair.watermark_kind OPERATOR(pg_catalog.=) 'event'
          AND pair.consumer_name OPERATOR(pg_catalog.=) NEW.consumer_name
          AND pair.consumer_version OPERATOR(pg_catalog.=) NEW.consumer_version
          AND pair.scope_kind OPERATOR(pg_catalog.=) NEW.scope_kind
          AND pair.scope_id OPERATOR(pg_catalog.=) NEW.scope_id
          AND pair.aggregate_type OPERATOR(pg_catalog.=) NEW.aggregate_type
          AND pair.aggregate_id OPERATOR(pg_catalog.=) NEW.aggregate_id
          AND pair.through_sequence OPERATOR(pg_catalog.=) NEW.event_sequence
          AND pair.causation_id OPERATOR(pg_catalog.=) NEW.event_id
          AND ROW(pair.tenant_id, pair.producer_group_id,
              pair.producer_group_generation, pair.producer_domain,
              pair.producer_generation, pair.writer_origin,
              pair.consumer_group_id, pair.consumer_group_generation,
              pair.consumer_domain, pair.consumer_generation)
          IS NOT DISTINCT FROM
              ROW(NEW.tenant_id, NEW.producer_group_id,
              NEW.producer_group_generation, NEW.producer_domain,
              NEW.producer_generation, NEW.writer_origin,
              NEW.consumer_group_id, NEW.consumer_group_generation,
              NEW.consumer_domain, NEW.consumer_generation)
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5635',
            MESSAGE = 'async inbox event receipt lacks its watermark';
    END IF;

    IF NEW.record_kind OPERATOR(pg_catalog.=) 'watermark' AND
       NEW.watermark_kind OPERATOR(pg_catalog.=) 'event' AND NOT EXISTS (
        SELECT 1 FROM platform.inbox AS pair
        WHERE pair.record_kind OPERATOR(pg_catalog.=) 'event_receipt'
          AND pair.consumer_name OPERATOR(pg_catalog.=) NEW.consumer_name
          AND pair.consumer_version OPERATOR(pg_catalog.=) NEW.consumer_version
          AND pair.scope_kind OPERATOR(pg_catalog.=) NEW.scope_kind
          AND pair.scope_id OPERATOR(pg_catalog.=) NEW.scope_id
          AND pair.aggregate_type OPERATOR(pg_catalog.=) NEW.aggregate_type
          AND pair.aggregate_id OPERATOR(pg_catalog.=) NEW.aggregate_id
          AND pair.event_sequence OPERATOR(pg_catalog.=) NEW.through_sequence
          AND pair.event_id OPERATOR(pg_catalog.=) NEW.causation_id
          AND ROW(pair.tenant_id, pair.producer_group_id,
              pair.producer_group_generation, pair.producer_domain,
              pair.producer_generation, pair.writer_origin,
              pair.consumer_group_id, pair.consumer_group_generation,
              pair.consumer_domain, pair.consumer_generation)
          IS NOT DISTINCT FROM
              ROW(NEW.tenant_id, NEW.producer_group_id,
              NEW.producer_group_generation, NEW.producer_domain,
              NEW.producer_generation, NEW.writer_origin,
              NEW.consumer_group_id, NEW.consumer_group_generation,
              NEW.consumer_domain, NEW.consumer_generation)
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5636',
            MESSAGE = 'async inbox event watermark lacks its receipt';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_require_inbox_event_pair_v1()
FROM PUBLIC;

CREATE CONSTRAINT TRIGGER inbox_event_watermark_pair_guard
AFTER INSERT ON platform.inbox
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_require_inbox_event_pair_v1();
ALTER TABLE platform.inbox
    ENABLE ALWAYS TRIGGER inbox_event_watermark_pair_guard;
