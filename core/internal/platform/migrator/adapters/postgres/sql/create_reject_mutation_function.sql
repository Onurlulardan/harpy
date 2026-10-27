CREATE FUNCTION platform.platform_migration_history_reject_mutation()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING
        ERRCODE = '55000',
        MESSAGE = 'platform migration history is immutable';
    RETURN NULL;
END;
$function$;
