-- +goose Up
CREATE TABLE work_checklist_items (
    id uuid NOT NULL PRIMARY KEY,
    item_text text NOT NULL,
    completed boolean NOT NULL,
    position bigint NOT NULL,
    task_id uuid NOT NULL,
    deleted_at timestamptz,
    row_version bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    principal_id uuid NOT NULL,
    principal_kind text NOT NULL,
    active text NOT NULL,
    document jsonb NOT NULL,
    tasks text NOT NULL
);

-- +goose Down
DROP TABLE work_checklist_items;
