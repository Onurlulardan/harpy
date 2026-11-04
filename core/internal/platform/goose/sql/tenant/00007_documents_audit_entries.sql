-- +goose Up
CREATE TABLE documents_audit_entries (
    audit_id uuid NOT NULL PRIMARY KEY,
    operation_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    action text NOT NULL,
    aggregate_kind text NOT NULL,
    aggregate_id uuid NOT NULL,
    aggregate_version bigint NOT NULL,
    occurred_at timestamptz,
    reason text NOT NULL
);

-- +goose Down
DROP TABLE documents_audit_entries;
