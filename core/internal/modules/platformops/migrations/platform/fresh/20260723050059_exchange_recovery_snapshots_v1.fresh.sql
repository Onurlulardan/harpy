-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_export_recovery_snapshot_v1(
    p_process text
)
RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_backup'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'forgecrud-v2:core-database-backup-executor'
       OR p_process IS DISTINCT FROM
          'forgecrud-v2:core-database-backup-executor' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='recovery snapshot export denied';
    END IF;
    RETURN pg_catalog.pg_export_snapshot();
END;
$function$;

CREATE FUNCTION platform.platformops_import_recovery_snapshot_v1(
    p_snapshot_id text
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_backup'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'forgecrud-v2:core-database-backup-executor'
       OR p_snapshot_id !~ '^[0-9]+-[0-9A-F]+-[0-9]+$' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='recovery snapshot import denied';
    END IF;
    PERFORM pg_catalog.pg_import_snapshot(p_snapshot_id);
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_export_recovery_snapshot_v1(text)
FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_import_recovery_snapshot_v1(text)
FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformops_export_recovery_snapshot_v1(text)
TO forgecrud_backup_reader;
GRANT EXECUTE ON FUNCTION platform.platformops_import_recovery_snapshot_v1(text)
TO forgecrud_backup_reader;
