-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_guard_offering_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS NOT DISTINCT FROM 'INSERT' THEN
        IF NEW.lifecycle <> 'active' OR NEW.revision <> 1 OR NEW.retired_at IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'billing offering initial state is invalid';
        END IF;
        RETURN NEW;
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'DELETE'
       OR OLD.lifecycle <> 'active'
       OR NEW.id IS DISTINCT FROM OLD.id
       OR NEW.authority_id IS DISTINCT FROM OLD.authority_id
       OR NEW.offering_key IS DISTINCT FROM OLD.offering_key
       OR NEW.created_at IS DISTINCT FROM OLD.created_at
       OR NEW.revision IS DISTINCT FROM OLD.revision + 1
       OR NOT (
           (NEW.lifecycle = 'active' AND NEW.retired_at IS NULL)
           OR (NEW.lifecycle = 'retired' AND NEW.retired_at >= OLD.created_at)
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing offering transition is invalid';
    END IF;
    RETURN NEW;
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_guard_offering_v1() FROM PUBLIC;

CREATE FUNCTION platform.billing_guard_discount_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS NOT DISTINCT FROM 'INSERT' THEN
        IF NEW.lifecycle <> 'active' OR NEW.revision <> 1 OR NEW.revoked_at IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'billing discount initial state is invalid';
        END IF;
        RETURN NEW;
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'DELETE'
       OR OLD.lifecycle <> 'active'
       OR NEW.id IS DISTINCT FROM OLD.id
       OR NEW.authority_id IS DISTINCT FROM OLD.authority_id
       OR NEW.normalized_code IS DISTINCT FROM OLD.normalized_code
       OR NEW.created_at IS DISTINCT FROM OLD.created_at
       OR NEW.revision IS DISTINCT FROM OLD.revision + 1
       OR NOT (
           (NEW.lifecycle = 'active' AND NEW.revoked_at IS NULL)
           OR (NEW.lifecycle = 'revoked' AND NEW.revoked_at >= OLD.created_at)
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing discount transition is invalid';
    END IF;
    RETURN NEW;
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_guard_discount_v1() FROM PUBLIC;

CREATE TRIGGER billing_offerings_transition_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_offerings
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_offering_v1();
ALTER TABLE platform.billing_offerings
    ENABLE ALWAYS TRIGGER billing_offerings_transition_guard;

CREATE TRIGGER billing_discounts_transition_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_discounts
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_discount_v1();
ALTER TABLE platform.billing_discounts
    ENABLE ALWAYS TRIGGER billing_discounts_transition_guard;
-- +goose StatementEnd

