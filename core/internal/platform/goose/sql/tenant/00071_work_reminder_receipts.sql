-- +goose Up
CREATE TABLE work_reminder_receipts (
    task_id uuid NOT NULL PRIMARY KEY,
    task_version bigint NOT NULL,
    due_at timestamptz,
    event_id uuid NOT NULL,
    dispatched_at timestamptz
);

-- +goose Down
DROP TABLE work_reminder_receipts;
