-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_guard_base_assignment_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS DISTINCT FROM 'INSERT' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing base offering assignment is immutable';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended(
            'billing:base-offering:' || NEW.tenant_id::text, 0
        )
    );
    IF EXISTS (
        SELECT 1
        FROM platform.billing_tenant_base_offerings AS assignment
        WHERE assignment.tenant_id = NEW.tenant_id
          AND assignment.effective_start <
              COALESCE(NEW.effective_end, 'infinity'::timestamptz)
          AND NEW.effective_start <
              COALESCE(assignment.effective_end, 'infinity'::timestamptz)
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23P01',
            CONSTRAINT = 'billing_tenant_base_offerings_no_overlap',
            MESSAGE = 'billing base offering assignments cannot overlap';
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_guard_base_assignment_v1() FROM PUBLIC;

CREATE FUNCTION platform.billing_guard_subscription_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    transition_allowed boolean;
BEGIN
    IF TG_OP IS NOT DISTINCT FROM 'INSERT' THEN
        IF NEW.revision IS DISTINCT FROM 1
           OR NEW.updated_at IS DISTINCT FROM NEW.created_at THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'billing subscription must begin at revision one';
        END IF;
        RETURN NEW;
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing subscription history cannot be deleted';
    END IF;
    IF OLD.state IN ('cancelled', 'expired') THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing terminal subscription cannot change';
    END IF;
    IF NEW.id IS DISTINCT FROM OLD.id
       OR NEW.authority_id IS DISTINCT FROM OLD.authority_id
       OR NEW.tenant_id IS DISTINCT FROM OLD.tenant_id
       OR NEW.provider_account_id IS DISTINCT FROM OLD.provider_account_id
       OR NEW.created_at IS DISTINCT FROM OLD.created_at
       OR NEW.revision IS DISTINCT FROM OLD.revision + 1
       OR NEW.updated_at <= OLD.updated_at THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing subscription revision is invalid';
    END IF;
    transition_allowed := NEW.state = OLD.state OR
        (OLD.state = 'pending' AND NEW.state IN ('trialing', 'active', 'cancelled')) OR
        (OLD.state = 'trialing' AND NEW.state IN (
            'active', 'past_due', 'cancelled', 'expired'
        )) OR
        (OLD.state = 'active' AND NEW.state IN (
            'past_due', 'paused', 'cancelled', 'expired'
        )) OR
        (OLD.state = 'past_due' AND NEW.state IN (
            'active', 'paused', 'cancelled', 'expired'
        )) OR
        (OLD.state = 'paused' AND NEW.state IN ('active', 'cancelled', 'expired'));
    IF NOT transition_allowed THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing subscription transition is forbidden';
    END IF;
    IF (NEW.offering_version_id IS DISTINCT FROM OLD.offering_version_id
        OR NEW.price_version_id IS DISTINCT FROM OLD.price_version_id)
       AND NOT EXISTS (
           SELECT 1 FROM platform.billing_subscription_changes AS change_row
           WHERE change_row.subscription_id = NEW.id
             AND change_row.after_offering_version_id = NEW.offering_version_id
             AND change_row.after_price_version_id = NEW.price_version_id
             AND change_row.state = 'applied'
             AND change_row.resolved_at <= NEW.updated_at
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing subscription price change has no applied decision';
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_guard_subscription_v1() FROM PUBLIC;

CREATE TRIGGER billing_tenant_base_offerings_interval_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_tenant_base_offerings
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_base_assignment_v1();
ALTER TABLE platform.billing_tenant_base_offerings
    ENABLE ALWAYS TRIGGER billing_tenant_base_offerings_interval_guard;

CREATE TRIGGER billing_subscriptions_transition_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_subscriptions
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_subscription_v1();
ALTER TABLE platform.billing_subscriptions
    ENABLE ALWAYS TRIGGER billing_subscriptions_transition_guard;
-- +goose StatementEnd
