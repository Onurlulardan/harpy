-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_reject_mutation()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING
        ERRCODE = '55000',
        MESSAGE = pg_catalog.format(
            '%s is forbidden for append-only tenancy relation %I.%I',
            TG_OP,
            TG_TABLE_SCHEMA,
            TG_TABLE_NAME
        );
END;
$function$;
-- +goose StatementEnd
