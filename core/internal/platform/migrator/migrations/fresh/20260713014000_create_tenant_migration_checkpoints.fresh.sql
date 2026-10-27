-- +goose Up
CREATE TABLE platform.tenant_migration_checkpoints (
    tenant_id uuid NOT NULL,
    owner text NOT NULL CHECK (owner ~ '^[a-z][a-z0-9_-]{0,63}$'),
    logical_version bigint NOT NULL CHECK (logical_version BETWEEN 20000101000000 AND 99991231235959),
    database_origin text NOT NULL CHECK (database_origin IN ('fresh', 'v1_adopted')),
    selected_path text NOT NULL CHECK (selected_path ~ '^[a-z0-9][a-z0-9_./-]*\.sql$'),
    selected_sha256 text NOT NULL CHECK (selected_sha256 ~ '^[0-9a-f]{64}$'),
    run_id uuid NOT NULL REFERENCES platform.migration_runs(run_id) ON DELETE RESTRICT,
    state text NOT NULL CHECK (state IN ('pending', 'running', 'succeeded', 'failed', 'fenced')),
    attempt integer NOT NULL CHECK (attempt > 0),
    started_at timestamptz,
    finished_at timestamptz,
    error_code text CHECK (error_code ~ '^[a-z][a-z0-9_]{2,63}$'),
    updated_at timestamptz NOT NULL DEFAULT statement_timestamp(),
    PRIMARY KEY (tenant_id, owner, logical_version),
    CHECK (position('..' IN selected_path) = 0 AND position('//' IN selected_path) = 0),
    CHECK ((state = 'pending' AND started_at IS NULL AND finished_at IS NULL AND error_code IS NULL)
        OR (state = 'running' AND started_at IS NOT NULL AND finished_at IS NULL AND error_code IS NULL)
        OR (state = 'succeeded' AND started_at IS NOT NULL AND finished_at IS NOT NULL AND error_code IS NULL)
        OR (state IN ('failed', 'fenced') AND finished_at IS NOT NULL AND error_code IS NOT NULL))
);
