-- +goose Up
GRANT USAGE ON SCHEMA platform TO forgecrud_runtime, forgecrud_backup_reader;

GRANT SELECT ON TABLE
    platform.environment_identity,
    platform.bootstrap_receipts,
    platform.extension_registry,
    platform.migration_runs,
    platform.schema_contracts,
    platform.database_generation_manifests,
    platform.release_generations
TO forgecrud_runtime, forgecrud_backup_reader;
