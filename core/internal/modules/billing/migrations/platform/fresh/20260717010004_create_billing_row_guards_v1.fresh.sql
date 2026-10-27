-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_guard_global_row_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    selected_authority uuid;
BEGIN
    IF TG_OP IS NOT DISTINCT FROM 'DELETE' THEN
        selected_authority := OLD.authority_id;
    ELSE
        selected_authority := NEW.authority_id;
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'UPDATE'
       AND NEW.authority_id IS DISTINCT FROM OLD.authority_id THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing global authority identity is immutable';
    END IF;
    PERFORM platform.billing_assert_global_authority_v1(selected_authority);
    IF TG_OP IS NOT DISTINCT FROM 'DELETE' THEN RETURN OLD; END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_guard_global_row_v1() FROM PUBLIC;

CREATE FUNCTION platform.billing_guard_tenant_row_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    selected_authority uuid;
    selected_tenant uuid;
BEGIN
    IF TG_OP IS NOT DISTINCT FROM 'DELETE' THEN
        selected_authority := OLD.authority_id;
        selected_tenant := OLD.tenant_id;
    ELSE
        selected_authority := NEW.authority_id;
        selected_tenant := NEW.tenant_id;
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'UPDATE' AND (
        NEW.authority_id IS DISTINCT FROM OLD.authority_id
        OR NEW.tenant_id IS DISTINCT FROM OLD.tenant_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing tenant authority identity is immutable';
    END IF;
    PERFORM platform.billing_assert_tenant_authority_v1(selected_authority);
    IF NOT EXISTS (
        SELECT 1 FROM platform.billing_tenant_authorities AS authority_row
        WHERE authority_row.authority_id = selected_authority
          AND authority_row.tenant_id = selected_tenant
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23503',
            MESSAGE = 'billing tenant authority identity does not match tenant';
    END IF;
    IF TG_OP IS NOT DISTINCT FROM 'DELETE' THEN RETURN OLD; END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_guard_tenant_row_v1() FROM PUBLIC;

CREATE FUNCTION platform.billing_guard_immutable_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS DISTINCT FROM 'INSERT' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing immutable history cannot be changed';
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_guard_immutable_v1() FROM PUBLIC;
-- +goose StatementEnd
