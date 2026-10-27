-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.cutover_reject_mutation()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING
        ERRCODE = '55000',
        MESSAGE = pg_catalog.format(
            '%s is forbidden for append-only cutover relation %I.%I',
            TG_OP,
            TG_TABLE_SCHEMA,
            TG_TABLE_NAME
        );
    RETURN OLD;
END;
$function$;
-- +goose StatementEnd
