-- +goose Up
CREATE TABLE platform.tenancy_tenant_migration_executor_readiness (
    profile_instance_id uuid PRIMARY KEY,
    release_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    deployment_mode text NOT NULL CHECK (deployment_mode IN ('onpremise', 'saas')),
    runtime_generation bigint NOT NULL CHECK (runtime_generation > 0),
    database_generation bigint NOT NULL CHECK (database_generation > 0),
    migration_manifest_sha256 text NOT NULL
        CHECK (migration_manifest_sha256 ~ '^[0-9a-f]{64}$'),
    state text NOT NULL CHECK (state IN ('active', 'retired')),
    ready_until timestamptz NOT NULL,
    heartbeat_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    retired_at timestamptz,
    CHECK (created_at <= heartbeat_at AND heartbeat_at < ready_until),
    CHECK ((state = 'active' AND retired_at IS NULL) OR
        (state = 'retired' AND retired_at IS NOT NULL))
);

CREATE TABLE platform.tenancy_tenant_migration_jobs (
    job_id uuid PRIMARY KEY,
    operation_id uuid NOT NULL UNIQUE
        REFERENCES platform.tenancy_onboarding_operations (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    origin text NOT NULL DEFAULT 'new' CHECK (origin = 'new'),
    release_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    deployment_mode text NOT NULL CHECK (deployment_mode IN ('onpremise', 'saas')),
    runtime_generation bigint NOT NULL CHECK (runtime_generation > 0),
    database_generation bigint NOT NULL CHECK (database_generation > 0),
    migration_manifest_sha256 text NOT NULL
        CHECK (migration_manifest_sha256 ~ '^[0-9a-f]{64}$'),
    state text NOT NULL CHECK (state IN (
        'pending', 'running', 'succeeded', 'failed',
        'reconcile_required', 'cleanup_required', 'fenced'
    )),
    current_attempt integer NOT NULL DEFAULT 0
        CHECK (current_attempt BETWEEN 0 AND 8),
    error_code text CHECK (error_code ~ '^[a-z][a-z0-9_]{2,63}$'),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    completed_at timestamptz,
    CHECK ((state IN ('pending', 'running') AND completed_at IS NULL) OR
        (state NOT IN ('pending', 'running') AND completed_at IS NOT NULL)),
    CHECK ((state IN ('pending', 'running', 'succeeded') AND error_code IS NULL) OR
        (state IN ('failed', 'reconcile_required', 'cleanup_required', 'fenced')
            AND error_code IS NOT NULL))
);

CREATE TABLE platform.tenancy_tenant_migration_attempts (
    job_id uuid NOT NULL
        REFERENCES platform.tenancy_tenant_migration_jobs (job_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    attempt integer NOT NULL CHECK (attempt BETWEEN 1 AND 8),
    migration_run_id uuid NOT NULL UNIQUE,
    profile_instance_id uuid NOT NULL,
    lease_id uuid NOT NULL UNIQUE,
    state text NOT NULL CHECK (state IN (
        'running', 'succeeded', 'failed', 'reconcile_required', 'cleanup_required'
    )),
    leased_at timestamptz NOT NULL,
    heartbeat_at timestamptz NOT NULL,
    lease_until timestamptz NOT NULL,
    finished_at timestamptz,
    error_code text CHECK (error_code ~ '^[a-z][a-z0-9_]{2,63}$'),
    PRIMARY KEY (job_id, attempt),
    CHECK (leased_at <= heartbeat_at AND heartbeat_at < lease_until),
    CHECK ((state = 'running' AND finished_at IS NULL AND error_code IS NULL) OR
        (state = 'succeeded' AND finished_at IS NOT NULL AND error_code IS NULL) OR
        (state IN ('failed', 'reconcile_required', 'cleanup_required')
            AND finished_at IS NOT NULL AND error_code IS NOT NULL))
);

CREATE TABLE platform.tenancy_tenant_migration_receipts (
    receipt_id uuid PRIMARY KEY,
    job_id uuid NOT NULL,
    attempt integer NOT NULL,
    receipt_kind text NOT NULL CHECK (receipt_kind IN (
        'schema_created', 'migration_completed', 'failure',
        'reconcile_required', 'cleanup_required', 'offline_cleanup_completed'
    )),
    migration_run_id uuid NOT NULL,
    receipt_sha256 text NOT NULL UNIQUE CHECK (receipt_sha256 ~ '^[0-9a-f]{64}$'),
    evidence_sha256 text NOT NULL CHECK (evidence_sha256 ~ '^[0-9a-f]{64}$'),
    occurred_at timestamptz NOT NULL,
    FOREIGN KEY (job_id, attempt)
        REFERENCES platform.tenancy_tenant_migration_attempts (job_id, attempt)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT tenancy_tenant_migration_receipt_attempt_kind_uq
        UNIQUE (job_id, attempt, receipt_kind)
);

CREATE UNIQUE INDEX tenancy_tenant_migration_schema_receipt_uq
ON platform.tenancy_tenant_migration_receipts (job_id)
WHERE receipt_kind = 'schema_created';

CREATE UNIQUE INDEX tenancy_tenant_migration_completion_receipt_uq
ON platform.tenancy_tenant_migration_receipts (job_id)
WHERE receipt_kind = 'migration_completed';

CREATE UNIQUE INDEX tenancy_tenant_migration_active_release_ready_uq
ON platform.tenancy_tenant_migration_executor_readiness (
    profile_instance_id, release_id, runtime_generation, database_generation
) WHERE state = 'active';
