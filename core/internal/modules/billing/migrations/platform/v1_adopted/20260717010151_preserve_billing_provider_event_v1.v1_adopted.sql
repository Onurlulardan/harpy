-- +goose Up
ALTER TABLE platform.billing_provider_events
    DROP CONSTRAINT billing_provider_events_receipt_fk,
    ALTER COLUMN webhook_receipt_id DROP NOT NULL,
    ADD CONSTRAINT billing_provider_events_receipt_fk FOREIGN KEY (
        tenant_id, webhook_receipt_id
    ) REFERENCES platform.billing_webhook_receipts (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE SET NULL (webhook_receipt_id);

-- +goose StatementBegin
CREATE FUNCTION platform.billing_guard_provider_event_retention_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS NOT DISTINCT FROM 'INSERT' THEN
        RETURN NEW;
    END IF;
    IF NOT platform.billing_retention_delete_allowed_v1(
        OLD.authority_id, OLD.tenant_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing provider event is immutable';
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'DELETE' THEN
        RETURN OLD;
    END IF;
    IF OLD.webhook_receipt_id IS NOT NULL
       AND NEW.webhook_receipt_id IS NULL
       AND (pg_catalog.to_jsonb(NEW) - 'webhook_receipt_id') =
           (pg_catalog.to_jsonb(OLD) - 'webhook_receipt_id') THEN
        RETURN NEW;
    END IF;
    RAISE EXCEPTION USING ERRCODE = '55000',
        MESSAGE = 'billing provider event retention update is forbidden';
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_guard_provider_event_retention_v1()
    FROM PUBLIC;

ALTER TABLE platform.billing_provider_events
    DISABLE TRIGGER billing_provider_events_immutable_guard;
CREATE TRIGGER billing_provider_events_retention_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_provider_events
FOR EACH ROW
EXECUTE FUNCTION platform.billing_guard_provider_event_retention_v1();
ALTER TABLE platform.billing_provider_events
    ENABLE ALWAYS TRIGGER billing_provider_events_retention_immutable_guard;
-- +goose StatementEnd
