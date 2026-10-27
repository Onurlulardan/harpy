-- +goose Up
ALTER TABLE platform.migration_runs
DROP CONSTRAINT migration_runs_command_check;

ALTER TABLE platform.migration_runs
ADD CONSTRAINT migration_runs_command_check CHECK (command IN (
    'plan', 'verify', 'up', 'schema_executor', 'tenant_migration_executor'
));
