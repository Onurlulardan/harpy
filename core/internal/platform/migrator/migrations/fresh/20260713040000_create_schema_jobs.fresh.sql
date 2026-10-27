-- +goose Up
CREATE TABLE platform.schema_jobs (
    job_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    owner text NOT NULL CHECK (owner ~ '^[a-z][a-z0-9_]{0,62}$'),
    resource_kind text NOT NULL
        CHECK (resource_kind IN ('form_storage', 'form_purge')),
    resource_id uuid NOT NULL,
    resource_revision bigint NOT NULL CHECK (resource_revision > 0),
    compiler_version text NOT NULL
        CHECK (compiler_version ~ '^[0-9]+\.[0-9]+\.[0-9]+$'),
    plan_contract_version text NOT NULL
        CHECK (plan_contract_version = 'forms-schema-plan.v1'),
    canonical_plan jsonb NOT NULL,
    plan_sha256 text NOT NULL CHECK (plan_sha256 ~ '^[0-9a-f]{64}$'),
    step_count integer NOT NULL CHECK (step_count BETWEEN 1 AND 4096),
    next_step_ordinal integer NOT NULL DEFAULT 0,
    generation bigint NOT NULL DEFAULT 1 CHECK (generation > 0),
    state text NOT NULL DEFAULT 'queued',
    next_executor text NOT NULL DEFAULT 'migrator',
    attempt integer NOT NULL DEFAULT 0 CHECK (attempt >= 0),
    available_at timestamptz NOT NULL,
    lease_id uuid,
    lease_owner_id uuid,
    leased_at timestamptz,
    lease_heartbeat_at timestamptz,
    lease_expires_at timestamptz,
    last_lease_expired_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT pg_catalog.statement_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT pg_catalog.statement_timestamp(),
    finished_at timestamptz,
    failure_code text CHECK (failure_code ~ '^[a-z][a-z0-9_]{2,63}$'),
    CONSTRAINT schema_jobs_tenant_fk
        FOREIGN KEY (tenant_id) REFERENCES platform.tenants (id) ON DELETE RESTRICT,
    CONSTRAINT schema_jobs_plan_unique UNIQUE (tenant_id, owner, plan_sha256),
    CONSTRAINT schema_jobs_hash_binding_unique UNIQUE (job_id, plan_sha256),
    CONSTRAINT schema_jobs_plan_shape_check CHECK (
        pg_catalog.jsonb_typeof(canonical_plan) = 'object' AND
        pg_catalog.octet_length(canonical_plan::text) BETWEEN 2 AND 1048576 AND
        canonical_plan::text !~*
            '"(sql|raw_sql|function_body|routine_body|client_identifier)"[[:space:]]*:'
    ),
    CONSTRAINT schema_jobs_state_check CHECK (state IN (
        'queued', 'leased', 'running', 'lease_expired', 'awaiting_runtime',
        'succeeded', 'failed', 'cancelled'
    )),
    CONSTRAINT schema_jobs_executor_check CHECK (
        (state = 'queued' AND next_executor = 'migrator') OR
        (state = 'awaiting_runtime' AND next_executor = 'forms_runtime') OR
        (state IN ('leased', 'running', 'lease_expired') AND
            next_executor IN ('migrator', 'forms_runtime')) OR
        (state IN ('succeeded', 'failed', 'cancelled') AND next_executor = 'none')
    ),
    CONSTRAINT schema_jobs_step_pointer_check CHECK (
        next_step_ordinal BETWEEN 0 AND step_count AND
        ((state = 'succeeded' AND next_step_ordinal = step_count) OR
         (state <> 'succeeded' AND next_step_ordinal < step_count))
    ),
    CONSTRAINT schema_jobs_lease_check CHECK (
        (state IN ('leased', 'running') AND
            lease_id IS NOT NULL AND lease_owner_id IS NOT NULL AND
            leased_at IS NOT NULL AND lease_heartbeat_at IS NOT NULL AND
            lease_expires_at IS NOT NULL AND
            leased_at <= lease_heartbeat_at AND lease_heartbeat_at <= updated_at AND
            updated_at < lease_expires_at) OR
        (state = 'lease_expired' AND
            lease_id IS NOT NULL AND lease_owner_id IS NOT NULL AND
            leased_at IS NOT NULL AND lease_heartbeat_at IS NOT NULL AND
            lease_expires_at IS NOT NULL AND last_lease_expired_at IS NOT NULL AND
            leased_at <= lease_heartbeat_at AND
            lease_expires_at <= last_lease_expired_at AND
            last_lease_expired_at <= updated_at) OR
        (state NOT IN ('leased', 'running', 'lease_expired') AND
            lease_id IS NULL AND lease_owner_id IS NULL AND leased_at IS NULL AND
            lease_heartbeat_at IS NULL AND lease_expires_at IS NULL)
    ),
    CONSTRAINT schema_jobs_outcome_check CHECK (
        (state IN ('queued', 'leased', 'running', 'awaiting_runtime') AND
            finished_at IS NULL AND failure_code IS NULL) OR
        (state = 'lease_expired' AND
            finished_at IS NULL AND failure_code = 'lease_expired') OR
        (state = 'succeeded' AND finished_at IS NOT NULL AND failure_code IS NULL) OR
        (state IN ('failed', 'cancelled') AND
            finished_at IS NOT NULL AND failure_code IS NOT NULL)
    ),
    CONSTRAINT schema_jobs_time_check CHECK (
        created_at <= updated_at AND created_at <= available_at AND
        (last_lease_expired_at IS NULL OR last_lease_expired_at <= updated_at) AND
        (finished_at IS NULL OR updated_at <= finished_at)
    )
);
