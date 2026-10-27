-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.cutover_reject_delete()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING
        ERRCODE = '55000',
        MESSAGE = pg_catalog.format(
            'delete is forbidden for immutable cutover relation %I.%I',
            TG_TABLE_SCHEMA,
            TG_TABLE_NAME
        );
    RETURN OLD;
END;
$function$;
-- +goose StatementEnd
