-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_validate_release_generation_finalization()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM platform.release_generations AS release_row
        JOIN platform.schema_contracts AS contract_row
          ON contract_row.component = 'core'
         AND contract_row.verified_run_id = release_row.migration_run_id
         AND contract_row.database_generation = release_row.database_generation
         AND contract_row.minimum_runtime_generation = release_row.minimum_runtime_generation
         AND contract_row.migration_manifest_sha256 = release_row.migration_manifest_sha256
         AND contract_row.extension_manifest_sha256 = release_row.extension_manifest_sha256
         AND contract_row.status = 'ready'
        WHERE release_row.migration_run_id = NEW.migration_run_id
          AND release_row.xmin::text = pg_catalog.pg_current_xact_id()::text
          AND contract_row.xmin::text = pg_catalog.pg_current_xact_id()::text
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'release generation is not the finalized active contract';
    END IF;
    RETURN NULL;
END;
$function$;
-- +goose StatementEnd
