-- +goose Up
CREATE TABLE notifications_messages (
    id uuid NOT NULL PRIMARY KEY,
    purged_at timestamptz,
    read_at timestamptz,
    row_version bigint NOT NULL,
    recipient_id uuid NOT NULL,
    template_version_id uuid NOT NULL,
    producer text NOT NULL,
    notification_type text NOT NULL,
    category text NOT NULL,
    template_kind text NOT NULL,
    locale text NOT NULL,
    subject text NOT NULL,
    body jsonb NOT NULL,
    data_digest text NOT NULL,
    subject_digest text NOT NULL,
    body_digest text NOT NULL,
    inbox_revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    dismissed_at timestamptz
);

-- +goose Down
DROP TABLE notifications_messages;
