-- +goose Up
CREATE TABLE platform.schema_job_steps (
    job_id uuid NOT NULL,
    ordinal integer NOT NULL CHECK (ordinal BETWEEN 0 AND 4095),
    plan_sha256 text NOT NULL CHECK (plan_sha256 ~ '^[0-9a-f]{64}$'),
    operation_id text NOT NULL,
    transaction_class text NOT NULL,
    executor_kind text NOT NULL,
    identifier_mapping jsonb NOT NULL,
    arguments jsonb NOT NULL,
    step_sha256 text NOT NULL CHECK (step_sha256 ~ '^[0-9a-f]{64}$'),
    generation bigint NOT NULL DEFAULT 1 CHECK (generation > 0),
    state text NOT NULL DEFAULT 'pending',
    next_executor text NOT NULL,
    attempt integer NOT NULL DEFAULT 0 CHECK (attempt >= 0),
    started_at timestamptz,
    finished_at timestamptz,
    error_code text CHECK (error_code ~ '^[a-z][a-z0-9_]{2,63}$'),
    created_at timestamptz NOT NULL DEFAULT pg_catalog.statement_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT pg_catalog.statement_timestamp(),
    PRIMARY KEY (job_id, ordinal),
    CONSTRAINT schema_job_steps_job_fk
        FOREIGN KEY (job_id, plan_sha256)
        REFERENCES platform.schema_jobs (job_id, plan_sha256) ON DELETE RESTRICT,
    CONSTRAINT schema_job_steps_hash_unique UNIQUE (
        job_id, ordinal, plan_sha256, step_sha256
    ),
    CONSTRAINT schema_job_steps_step_hash_unique UNIQUE (job_id, step_sha256),
    CONSTRAINT schema_job_steps_operation_check CHECK (operation_id IN (
        'activate_generation', 'add_constraint', 'add_field_column',
        'add_relation_table', 'backfill_field_batch', 'cleanup_invalid_index',
        'create_form_table', 'create_index', 'create_index_concurrently',
        'disable_old_writer', 'drop_constraint', 'drop_field_column',
        'drop_form_table', 'drop_index', 'drop_index_concurrently',
        'drop_relation_table', 'enable_dual_write', 'retain_generation',
        'validate_constraint', 'verify_storage'
    )),
    CONSTRAINT schema_job_steps_executor_kind_check CHECK (
        (executor_kind = 'forms_runtime' AND operation_id IN (
            'activate_generation', 'disable_old_writer',
            'enable_dual_write', 'retain_generation'
        )) OR
        (executor_kind = 'migrator' AND operation_id IN (
            'add_constraint', 'add_field_column', 'add_relation_table',
            'backfill_field_batch', 'cleanup_invalid_index', 'create_form_table',
            'create_index', 'create_index_concurrently', 'drop_constraint',
            'drop_field_column', 'drop_form_table', 'drop_index',
            'drop_index_concurrently', 'drop_relation_table',
            'validate_constraint', 'verify_storage'
        ))
    ),
    CONSTRAINT schema_job_steps_transaction_class_check CHECK (
        (operation_id IN (
            'cleanup_invalid_index', 'create_index_concurrently',
            'drop_index_concurrently'
        ) AND transaction_class = 'concurrent_index') OR
        (operation_id NOT IN (
            'cleanup_invalid_index', 'create_index_concurrently',
            'drop_index_concurrently'
        ) AND transaction_class = 'transactional')
    ),
    CONSTRAINT schema_job_steps_payload_check CHECK (
        pg_catalog.jsonb_typeof(identifier_mapping) = 'object' AND
        identifier_mapping <> '{}'::jsonb AND
        pg_catalog.octet_length(identifier_mapping::text) BETWEEN 2 AND 65536 AND
        pg_catalog.jsonb_typeof(arguments) = 'object' AND
        pg_catalog.octet_length(arguments::text) BETWEEN 2 AND 262144 AND
        (identifier_mapping::text || arguments::text) !~*
            '"(sql|raw_sql|function_body|routine_body|client_identifier)"[[:space:]]*:'
    ),
    CONSTRAINT schema_job_steps_state_check CHECK (state IN (
        'pending', 'running', 'retryable', 'lease_expired', 'succeeded',
        'failed', 'blocked', 'skipped', 'cancelled'
    )),
    CONSTRAINT schema_job_steps_state_shape_check CHECK (
        (state = 'pending' AND attempt = 0 AND next_executor = executor_kind AND
            started_at IS NULL AND finished_at IS NULL AND error_code IS NULL) OR
        (state = 'running' AND attempt > 0 AND next_executor = executor_kind AND
            started_at IS NOT NULL AND finished_at IS NULL AND error_code IS NULL) OR
        (state IN ('retryable', 'lease_expired') AND attempt > 0 AND
            next_executor = executor_kind AND started_at IS NOT NULL AND
            finished_at IS NOT NULL AND error_code IS NOT NULL) OR
        (state IN ('succeeded', 'skipped') AND attempt > 0 AND
            next_executor = 'none' AND started_at IS NOT NULL AND
            finished_at IS NOT NULL AND error_code IS NULL) OR
        (state IN ('failed', 'blocked', 'cancelled') AND attempt > 0 AND
            next_executor = 'none' AND started_at IS NOT NULL AND
            finished_at IS NOT NULL AND error_code IS NOT NULL)
    ),
    CONSTRAINT schema_job_steps_time_check CHECK (
        created_at <= updated_at AND
        (started_at IS NULL OR created_at <= started_at) AND
        (finished_at IS NULL OR started_at <= finished_at) AND
        (finished_at IS NULL OR finished_at <= updated_at)
    )
);
