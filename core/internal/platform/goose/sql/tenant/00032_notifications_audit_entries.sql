-- +goose Up
CREATE TABLE notifications_audit_entries (
    id uuid NOT NULL PRIMARY KEY,
    action text NOT NULL,
    resource_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    resource_revision bigint NOT NULL,
    occurred_at timestamptz
);

-- +goose Down
DROP TABLE notifications_audit_entries;
