-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_guard_usage_total_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS NOT DISTINCT FROM 'INSERT' THEN
        IF NEW.revision IS DISTINCT FROM 1 THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'billing usage total must begin at revision one';
        END IF;
        RETURN NEW;
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'DELETE'
       OR NEW.authority_id IS DISTINCT FROM OLD.authority_id
       OR NEW.tenant_id IS DISTINCT FROM OLD.tenant_id
       OR NEW.meter_period_id IS DISTINCT FROM OLD.meter_period_id
       OR NEW.entitlement_key IS DISTINCT FROM OLD.entitlement_key
       OR NEW.meter_kind IS DISTINCT FROM OLD.meter_kind
       OR NEW.revision IS DISTINCT FROM OLD.revision + 1
       OR NEW.updated_at <= OLD.updated_at THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing usage total revision is invalid';
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_guard_usage_total_v1() FROM PUBLIC;

CREATE FUNCTION platform.billing_guard_usage_reservation_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    transition_allowed boolean;
BEGIN
    IF TG_OP IS NOT DISTINCT FROM 'INSERT' THEN
        IF NEW.state IS DISTINCT FROM 'reserved'
           OR NEW.revision IS DISTINCT FROM 1
           OR NEW.resolved_at IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'billing usage reservation must begin reserved';
        END IF;
        RETURN NEW;
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'DELETE'
       OR NEW.id IS DISTINCT FROM OLD.id
       OR NEW.authority_id IS DISTINCT FROM OLD.authority_id
       OR NEW.tenant_id IS DISTINCT FROM OLD.tenant_id
       OR NEW.entitlement_key IS DISTINCT FROM OLD.entitlement_key
       OR NEW.meter_period_id IS DISTINCT FROM OLD.meter_period_id
       OR NEW.source_operation_id IS DISTINCT FROM OLD.source_operation_id
       OR NEW.amount IS DISTINCT FROM OLD.amount
       OR NEW.expires_at IS DISTINCT FROM OLD.expires_at
       OR NEW.created_at IS DISTINCT FROM OLD.created_at
       OR NEW.revision IS DISTINCT FROM OLD.revision + 1 THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing usage reservation identity is immutable';
    END IF;
    transition_allowed :=
        (OLD.state = 'reserved' AND NEW.state = 'reserved'
            AND OLD.effect_started_at IS NULL
            AND NEW.effect_started_at IS NOT NULL
            AND NEW.effect_receipt_id IS NOT NULL)
        OR (OLD.state = 'reserved'
            AND NEW.state IN ('committed', 'released', 'reconcile_required'))
        OR (OLD.state = 'reconcile_required'
            AND NEW.state IN ('committed', 'released'));
    IF NOT transition_allowed THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing usage reservation transition is forbidden';
    END IF;
    IF OLD.effect_started_at IS NOT NULL
       AND NEW.effect_started_at IS DISTINCT FROM OLD.effect_started_at THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing usage effect start is immutable';
    END IF;
    IF OLD.effect_receipt_id IS NOT NULL
       AND NEW.effect_receipt_id IS DISTINCT FROM OLD.effect_receipt_id THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing usage effect receipt is immutable';
    END IF;
    IF OLD.resolution_receipt_id IS NOT NULL
       AND NEW.resolution_receipt_id IS DISTINCT FROM OLD.resolution_receipt_id THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing usage resolution receipt is immutable';
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_guard_usage_reservation_v1() FROM PUBLIC;

CREATE TRIGGER billing_usage_totals_revision_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_usage_totals
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_usage_total_v1();
ALTER TABLE platform.billing_usage_totals
    ENABLE ALWAYS TRIGGER billing_usage_totals_revision_guard;

CREATE TRIGGER billing_usage_reservations_transition_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_usage_reservations
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_usage_reservation_v1();
ALTER TABLE platform.billing_usage_reservations
    ENABLE ALWAYS TRIGGER billing_usage_reservations_transition_guard;
-- +goose StatementEnd
