-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_reject_mutation()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING
        ERRCODE = '55000',
        MESSAGE = 'harpy immutable foundation evidence cannot be mutated';
END;
$function$;
-- +goose StatementEnd
