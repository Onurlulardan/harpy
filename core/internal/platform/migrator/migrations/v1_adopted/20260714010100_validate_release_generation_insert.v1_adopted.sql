-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_validate_release_generation_insert()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM platform.release_generations AS stored
        WHERE stored.migration_run_id = NEW.migration_run_id
          AND stored.release_id = NEW.release_id
          AND stored.environment_id = NEW.environment_id
          AND stored.database_origin = NEW.database_origin
          AND stored.database_generation = NEW.database_generation
          AND stored.minimum_runtime_generation = NEW.minimum_runtime_generation
          AND stored.runtime_generation = NEW.runtime_generation
          AND stored.minimum_database_generation = NEW.minimum_database_generation
          AND stored.migration_manifest_sha256 = NEW.migration_manifest_sha256
          AND stored.extension_manifest_sha256 = NEW.extension_manifest_sha256
          AND stored.postgres_image_digest = NEW.postgres_image_digest
          AND stored.bootstrap_receipt_run_id = NEW.bootstrap_receipt_run_id
          AND stored.extension_receipt_run_id = NEW.extension_receipt_run_id
          AND stored.cleanup_receipt_run_id = NEW.cleanup_receipt_run_id
    ) THEN
        RETURN NEW;
    END IF;
    IF EXISTS (
        SELECT 1 FROM platform.release_generations AS stored
        WHERE stored.migration_run_id = NEW.migration_run_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'release generation replay differs';
    END IF;
    IF EXISTS (
        SELECT 1 FROM platform.release_generations AS stored
        WHERE stored.environment_id = NEW.environment_id
          AND stored.release_id = NEW.release_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'release generation release fork';
    END IF;
    IF EXISTS (
        SELECT 1 FROM platform.release_generations AS stored
        WHERE stored.environment_id = NEW.environment_id
          AND stored.runtime_generation = NEW.runtime_generation
          AND stored.minimum_database_generation = NEW.minimum_database_generation
          AND stored.migration_manifest_sha256 = NEW.migration_manifest_sha256
          AND stored.extension_manifest_sha256 = NEW.extension_manifest_sha256
          AND stored.postgres_image_digest = NEW.postgres_image_digest
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'release generation binary tuple is duplicated';
    END IF;
    IF NOT EXISTS (
        SELECT 1
        FROM platform.migration_runs AS run_row
        JOIN platform.environment_identity AS identity_row
          ON identity_row.singleton = TRUE
         AND identity_row.environment_id = run_row.environment_id
         AND identity_row.database_origin = run_row.database_origin
        JOIN platform.bootstrap_receipts AS bootstrap_row
          ON bootstrap_row.kind = NEW.bootstrap_receipt_kind
         AND bootstrap_row.run_id = NEW.bootstrap_receipt_run_id
        JOIN platform.bootstrap_receipts AS extension_row
          ON extension_row.kind = NEW.extension_receipt_kind
         AND extension_row.run_id = NEW.extension_receipt_run_id
        JOIN platform.bootstrap_receipts AS cleanup_row
          ON cleanup_row.kind = NEW.cleanup_receipt_kind
         AND cleanup_row.run_id = NEW.cleanup_receipt_run_id
        WHERE run_row.run_id = NEW.migration_run_id
          AND run_row.release_id = NEW.release_id
          AND run_row.environment_id = NEW.environment_id
          AND run_row.database_origin = NEW.database_origin
          AND run_row.command = 'up'
          AND run_row.state = 'succeeded'
          AND run_row.migration_manifest_sha256 = NEW.migration_manifest_sha256
          AND run_row.extension_manifest_sha256 = NEW.extension_manifest_sha256
          AND run_row.xmin::text = pg_catalog.pg_current_xact_id()::text
          AND bootstrap_row.release_id = NEW.release_id
          AND extension_row.release_id = NEW.release_id
          AND cleanup_row.release_id = NEW.release_id
          AND bootstrap_row.environment_id = NEW.environment_id
          AND extension_row.environment_id = NEW.environment_id
          AND cleanup_row.environment_id = NEW.environment_id
          AND bootstrap_row.database_origin = NEW.database_origin
          AND extension_row.database_origin = NEW.database_origin
          AND cleanup_row.database_origin = NEW.database_origin
          AND bootstrap_row.database_system_identifier = identity_row.database_system_identifier
          AND extension_row.database_system_identifier = identity_row.database_system_identifier
          AND cleanup_row.database_system_identifier = identity_row.database_system_identifier
          AND bootstrap_row.database_oid = identity_row.database_oid
          AND extension_row.database_oid = identity_row.database_oid
          AND cleanup_row.database_oid = identity_row.database_oid
          AND bootstrap_row.xmin::text = pg_catalog.pg_current_xact_id()::text
          AND extension_row.xmin::text = pg_catalog.pg_current_xact_id()::text
          AND cleanup_row.xmin::text = pg_catalog.pg_current_xact_id()::text
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'release generation is outside exact finalization';
    END IF;
    IF NOT EXISTS (
        SELECT 1
        FROM platform.platform_migration_history AS history_row
        WHERE history_row.record_kind = 'migration'
          AND history_row.is_applied = TRUE
          AND history_row.version = (
              SELECT pg_catalog.max(latest_row.version)
              FROM platform.platform_migration_history AS latest_row
              WHERE latest_row.record_kind = 'migration'
                AND latest_row.is_applied = TRUE
          )
          AND history_row.target_database_generation = NEW.database_generation
          AND history_row.next_minimum_runtime_generation = NEW.minimum_runtime_generation
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'release generation differs from migration history';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd
