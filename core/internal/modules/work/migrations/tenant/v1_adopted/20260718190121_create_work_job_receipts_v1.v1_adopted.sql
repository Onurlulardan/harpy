-- +goose Up
CREATE TABLE work_job_receipts (
    job_id uuid PRIMARY KEY,
    job_kind text NOT NULL,
    payload_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(payload_sha256) = 32),
    lease_generation bigint NOT NULL CHECK (lease_generation > 0),
    state text NOT NULL CHECK (state IN ('RUNNING','SUCCEEDED')),
    outcome_sha256 bytea CHECK (outcome_sha256 IS NULL OR
        pg_catalog.octet_length(outcome_sha256) = 32),
    started_at timestamptz NOT NULL,
    completed_at timestamptz,
    CHECK ((state = 'RUNNING' AND outcome_sha256 IS NULL AND completed_at IS NULL) OR
           (state = 'SUCCEEDED' AND outcome_sha256 IS NOT NULL AND completed_at IS NOT NULL))
);

CREATE TABLE work_reminder_receipts (
    task_id uuid NOT NULL,
    task_version bigint NOT NULL CHECK (task_version > 0),
    due_at timestamptz NOT NULL,
    event_id uuid NOT NULL UNIQUE,
    dispatched_at timestamptz NOT NULL,
    PRIMARY KEY (task_id, task_version, due_at),
    CONSTRAINT work_reminder_receipts_task_fk FOREIGN KEY (task_id)
        REFERENCES work_tasks(id) ON UPDATE NO ACTION ON DELETE RESTRICT
);

GRANT SELECT ON work_job_receipts, work_reminder_receipts TO harpy_runtime;
GRANT INSERT (job_id, job_kind, payload_sha256, lease_generation, state,
    outcome_sha256, started_at, completed_at)
ON work_job_receipts TO harpy_runtime;
GRANT UPDATE (state, outcome_sha256, completed_at)
ON work_job_receipts TO harpy_runtime;
GRANT INSERT (task_id, task_version, due_at, event_id, dispatched_at)
ON work_reminder_receipts TO harpy_runtime;
