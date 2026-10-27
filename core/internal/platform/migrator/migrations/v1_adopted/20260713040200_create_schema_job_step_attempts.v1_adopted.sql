-- +goose Up
CREATE TABLE platform.schema_job_step_attempts (
    job_id uuid NOT NULL,
    step_ordinal integer NOT NULL,
    attempt integer NOT NULL CHECK (attempt > 0),
    plan_sha256 text NOT NULL CHECK (plan_sha256 ~ '^[0-9a-f]{64}$'),
    step_sha256 text NOT NULL CHECK (step_sha256 ~ '^[0-9a-f]{64}$'),
    lease_id uuid NOT NULL,
    lease_owner_id uuid NOT NULL,
    executor_kind text NOT NULL
        CHECK (executor_kind IN ('migrator', 'forms_runtime')),
    state text NOT NULL CHECK (state IN (
        'succeeded', 'retryable', 'lease_expired', 'failed',
        'blocked', 'skipped', 'cancelled'
    )),
    started_at timestamptz NOT NULL,
    finished_at timestamptz NOT NULL,
    outcome_sha256 text NOT NULL CHECK (outcome_sha256 ~ '^[0-9a-f]{64}$'),
    error_code text CHECK (error_code ~ '^[a-z][a-z0-9_]{2,63}$'),
    redacted_detail_sha256 text
        CHECK (redacted_detail_sha256 ~ '^[0-9a-f]{64}$'),
    recorded_at timestamptz NOT NULL DEFAULT pg_catalog.statement_timestamp(),
    PRIMARY KEY (job_id, step_ordinal, attempt),
    CONSTRAINT schema_job_step_attempts_step_fk
        FOREIGN KEY (job_id, step_ordinal, plan_sha256, step_sha256)
        REFERENCES platform.schema_job_steps (
            job_id, ordinal, plan_sha256, step_sha256
        ) ON DELETE RESTRICT,
    CONSTRAINT schema_job_step_attempts_outcome_check CHECK (
        (state IN ('succeeded', 'skipped') AND
            error_code IS NULL AND redacted_detail_sha256 IS NULL) OR
        (state IN (
            'retryable', 'lease_expired', 'failed', 'blocked', 'cancelled'
        ) AND error_code IS NOT NULL AND redacted_detail_sha256 IS NOT NULL)
    ),
    CONSTRAINT schema_job_step_attempts_time_check CHECK (
        started_at <= finished_at AND finished_at <= recorded_at
    )
);
