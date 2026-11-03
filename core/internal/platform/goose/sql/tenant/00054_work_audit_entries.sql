-- +goose Up
CREATE TABLE work_audit_entries (
    audit_id uuid NOT NULL PRIMARY KEY,
    operation_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    action text NOT NULL,
    aggregate_kind text NOT NULL,
    aggregate_id uuid NOT NULL,
    aggregate_version bigint NOT NULL,
    event_sequence text NOT NULL,
    reason text NOT NULL,
    occurred_at timestamptz
);

-- +goose Down
DROP TABLE work_audit_entries;
