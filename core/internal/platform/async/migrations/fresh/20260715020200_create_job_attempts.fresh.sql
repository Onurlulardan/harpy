-- +goose Up
CREATE TABLE platform.job_attempts (
    attempt_id uuid PRIMARY KEY,
    job_id uuid NOT NULL,
    scope_kind text NOT NULL CHECK (scope_kind IN ('global', 'tenant')),
    scope_id uuid NOT NULL,
    attempt integer NOT NULL CHECK (attempt BETWEEN 1 AND 25),
    lease_generation bigint NOT NULL CHECK (lease_generation > 0),
    lease_owner_id uuid NOT NULL,
    started_at timestamptz NOT NULL,
    finished_at timestamptz,
    outcome text CHECK (outcome IN (
        'succeeded', 'requeued', 'dead', 'cancelled', 'lease_expired'
    )),
    error_class text CHECK (error_class IN (
        'transient', 'permanent', 'conflict', 'cancelled'
    )),
    error_code text CHECK (error_code ~ '^[A-Z][A-Z0-9_]{1,63}$'),
    redacted_error_sha256 text
        CHECK (redacted_error_sha256 ~ '^[0-9a-f]{64}$'),
    effect_started_at timestamptz,
    effect_id uuid,
    effect_request_sha256 text
        CHECK (effect_request_sha256 ~ '^[0-9a-f]{64}$'),
    effect_receipt_id uuid,
    CONSTRAINT job_attempts_job_attempt_unique UNIQUE (job_id, attempt),
    CONSTRAINT job_attempts_job_fk
        FOREIGN KEY (scope_kind, scope_id, job_id)
        REFERENCES platform.jobs (scope_kind, scope_id, id) ON DELETE RESTRICT,
    CONSTRAINT job_attempts_lease_check CHECK (lease_generation = attempt),
    CONSTRAINT job_attempts_terminal_check CHECK (
        (finished_at IS NULL AND outcome IS NULL AND error_class IS NULL AND
            error_code IS NULL AND redacted_error_sha256 IS NULL AND
            effect_receipt_id IS NULL) OR
        (finished_at IS NOT NULL AND outcome IS NOT NULL)
    ),
    CONSTRAINT job_attempts_error_check CHECK (
        (outcome IS NULL AND error_class IS NULL AND error_code IS NULL AND
            redacted_error_sha256 IS NULL) OR
        (outcome = 'succeeded' AND error_class IS NULL AND error_code IS NULL AND
            redacted_error_sha256 IS NULL) OR
        (outcome = 'dead' AND error_class IS NULL AND
            error_code = 'JOB_OUTCOME_UNKNOWN' AND redacted_error_sha256 IS NULL AND
            effect_started_at IS NOT NULL AND effect_receipt_id IS NULL) OR
        (outcome IN (
            'requeued', 'dead', 'cancelled', 'lease_expired'
        ) AND error_class IS NOT NULL AND error_code IS NOT NULL AND
            redacted_error_sha256 IS NOT NULL)
    ),
    CONSTRAINT job_attempts_time_check CHECK (
        (effect_started_at IS NULL OR started_at <= effect_started_at) AND
        (finished_at IS NULL OR started_at <= finished_at) AND
        (effect_started_at IS NULL OR finished_at IS NULL OR
            effect_started_at <= finished_at)
    ),
    CONSTRAINT job_attempts_effect_identity_check CHECK (
        (effect_started_at IS NULL AND effect_id IS NULL AND
            effect_request_sha256 IS NULL) OR
        (effect_started_at IS NOT NULL AND effect_id IS NOT NULL AND
            effect_request_sha256 IS NOT NULL)
    )
);
