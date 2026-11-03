-- +goose Up
CREATE TABLE work_job_receipts (
    state text NOT NULL PRIMARY KEY,
    outcome_sha256 text NOT NULL,
    completed_at timestamptz,
    job_id uuid NOT NULL,
    job_kind text NOT NULL,
    payload_sha256 text NOT NULL,
    lease_generation bigint NOT NULL,
    started_at timestamptz
);

-- +goose Down
DROP TABLE work_job_receipts;
