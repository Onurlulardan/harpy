-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION tenancy_reject_schema_identity_mutation()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING
        ERRCODE = '55000', MESSAGE = 'tenant schema identity is immutable';
END;
$function$;
-- +goose StatementEnd
