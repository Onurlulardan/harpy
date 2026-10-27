-- +goose Up
GRANT USAGE ON SCHEMA platform TO harpy_runtime, harpy_backup_reader;

GRANT SELECT ON TABLE
    platform.environment_identity,
    platform.bootstrap_receipts,
    platform.extension_registry,
    platform.migration_runs,
    platform.schema_contracts,
    platform.database_generation_manifests,
    platform.release_generations
TO harpy_runtime, harpy_backup_reader;
