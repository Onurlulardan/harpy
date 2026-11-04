-- +goose Up
CREATE TABLE platform.job_attempts (
    attempt_id uuid NOT NULL,
    job_id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    attempt integer NOT NULL,
    lease_generation bigint NOT NULL,
    lease_owner_id uuid NOT NULL,
    started_at timestamptz NOT NULL,
    finished_at timestamptz,
    outcome text,
    error_class text,
    error_code text,
    redacted_error_sha256 text,
    effect_started_at timestamptz,
    effect_id uuid,
    effect_request_sha256 text,
    effect_receipt_id uuid
);

-- +goose Down
DROP TABLE platform.job_attempts;
