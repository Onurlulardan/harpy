-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_guard_tenant_registry()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        IF NEW.id IS DISTINCT FROM OLD.id
            OR NEW.slug IS DISTINCT FROM OLD.slug
            OR NEW.schema_name IS DISTINCT FROM OLD.schema_name
            OR NEW.created_at IS DISTINCT FROM OLD.created_at THEN
            RAISE EXCEPTION USING
                ERRCODE = '55000',
                MESSAGE = 'tenant id, slug, schema name, and creation time are immutable';
        END IF;
        RETURN NEW;
    END IF;
    RAISE EXCEPTION USING
        ERRCODE = '55000',
        MESSAGE = 'tenant registry hard delete is forbidden';
END;
$function$;
-- +goose StatementEnd
