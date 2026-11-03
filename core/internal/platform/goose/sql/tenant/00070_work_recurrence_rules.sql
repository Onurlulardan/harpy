-- +goose Up
CREATE TABLE work_recurrence_rules (
    id uuid NOT NULL PRIMARY KEY,
    lifecycle_state text NOT NULL,
    current_revision bigint NOT NULL,
    owner_event_id uuid NOT NULL,
    row_version bigint NOT NULL,
    updated_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE work_recurrence_rules;
