-- +goose Up
CREATE TABLE work_task_assignees (
    principal_id uuid NOT NULL PRIMARY KEY,
    task_id uuid NOT NULL,
    active text NOT NULL,
    principal_kind text NOT NULL,
    removed_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE work_task_assignees;
