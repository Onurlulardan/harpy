-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_retention_delete_allowed_v1(
    selected_authority uuid,
    selected_tenant uuid
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    RETURN SESSION_USER IS NOT DISTINCT FROM 'forgecrud_v2_runtime'
       AND selected_authority IS NOT NULL
       AND selected_tenant IS NOT NULL
       AND EXISTS (
           SELECT 1
           FROM platform.billing_retention_contexts AS retention_context
           WHERE retention_context.backend_pid = pg_catalog.pg_backend_pid()
             AND retention_context.transaction_id =
                 pg_catalog.pg_current_xact_id_if_assigned()
             AND retention_context.authority_id = selected_authority
             AND retention_context.tenant_id = selected_tenant
       );
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_retention_delete_allowed_v1(uuid, uuid)
    FROM PUBLIC;

CREATE FUNCTION platform.billing_guard_immutable_retention_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    previous_row jsonb;
    selected_authority uuid;
    selected_tenant uuid;
BEGIN
    IF TG_OP IS NOT DISTINCT FROM 'INSERT' THEN
        RETURN NEW;
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'DELETE' THEN
        previous_row := pg_catalog.to_jsonb(OLD);
        selected_authority := (previous_row ->> 'authority_id')::uuid;
        selected_tenant := (previous_row ->> 'tenant_id')::uuid;
        IF platform.billing_retention_delete_allowed_v1(
            selected_authority, selected_tenant
        ) THEN
            RETURN OLD;
        END IF;
    END IF;
    RAISE EXCEPTION USING ERRCODE = '55000',
        MESSAGE = 'billing immutable retention change is forbidden';
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_guard_immutable_retention_v1()
    FROM PUBLIC;

CREATE FUNCTION platform.billing_guard_legal_hold_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS NOT DISTINCT FROM 'INSERT' THEN
        IF NEW.state IS DISTINCT FROM 'active' OR NEW.revision IS DISTINCT FROM 1 THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'billing legal hold must begin active';
        END IF;
        RETURN NEW;
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'DELETE'
       OR OLD.state IS DISTINCT FROM 'active'
       OR NEW.state IS DISTINCT FROM 'released'
       OR NEW.id IS DISTINCT FROM OLD.id
       OR NEW.authority_id IS DISTINCT FROM OLD.authority_id
       OR NEW.tenant_id IS DISTINCT FROM OLD.tenant_id
       OR NEW.placed_by IS DISTINCT FROM OLD.placed_by
       OR NEW.place_operation_id IS DISTINCT FROM OLD.place_operation_id
       OR NEW.place_idempotency_key IS DISTINCT FROM OLD.place_idempotency_key
       OR NEW.place_request_sha256 IS DISTINCT FROM OLD.place_request_sha256
       OR NEW.reason IS DISTINCT FROM OLD.reason
       OR NEW.ticket_reference IS DISTINCT FROM OLD.ticket_reference
       OR NEW.placed_at IS DISTINCT FROM OLD.placed_at
       OR NEW.revision IS DISTINCT FROM OLD.revision + 1 THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing legal hold lifecycle is immutable';
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_guard_legal_hold_v1() FROM PUBLIC;

CREATE FUNCTION platform.billing_guard_retention_checkpoint_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS NOT DISTINCT FROM 'INSERT' THEN
        IF NEW.revision IS DISTINCT FROM 1 OR NEW.cycle_count IS DISTINCT FROM 1 THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'billing retention checkpoint must begin at one';
        END IF;
        RETURN NEW;
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'DELETE'
       OR NEW.authority_id IS DISTINCT FROM OLD.authority_id
       OR NEW.tenant_id IS DISTINCT FROM OLD.tenant_id
       OR NEW.revision IS DISTINCT FROM OLD.revision + 1
       OR NEW.cycle_count IS DISTINCT FROM OLD.cycle_count + 1
       OR NEW.raw_cutoff <= OLD.raw_cutoff
       OR NEW.entitlement_cutoff <= OLD.entitlement_cutoff
       OR NEW.financial_cutoff <= OLD.financial_cutoff
       OR NEW.last_completed_at <= OLD.last_completed_at THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing retention checkpoint revision is invalid';
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_guard_retention_checkpoint_v1()
    FROM PUBLIC;

CREATE TRIGGER billing_legal_holds_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_legal_holds
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_legal_holds
    ENABLE ALWAYS TRIGGER billing_legal_holds_authority_guard;
CREATE TRIGGER billing_legal_holds_lifecycle_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_legal_holds
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_legal_hold_v1();
ALTER TABLE platform.billing_legal_holds
    ENABLE ALWAYS TRIGGER billing_legal_holds_lifecycle_guard;

CREATE TRIGGER billing_retention_checkpoints_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_retention_checkpoints
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_retention_checkpoints
    ENABLE ALWAYS TRIGGER billing_retention_checkpoints_authority_guard;
CREATE TRIGGER billing_retention_checkpoints_revision_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_retention_checkpoints
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_retention_checkpoint_v1();
ALTER TABLE platform.billing_retention_checkpoints
    ENABLE ALWAYS TRIGGER billing_retention_checkpoints_revision_guard;

CREATE TRIGGER billing_retention_receipts_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_retention_receipts
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_retention_receipts
    ENABLE ALWAYS TRIGGER billing_retention_receipts_authority_guard;
CREATE TRIGGER billing_retention_receipts_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_retention_receipts
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_retention_receipts
    ENABLE ALWAYS TRIGGER billing_retention_receipts_immutable_guard;
-- +goose StatementEnd
