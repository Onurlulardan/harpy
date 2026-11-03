-- +goose Up
CREATE TABLE notifications_realtime_hints (
    message_id uuid NOT NULL,
    id uuid NOT NULL PRIMARY KEY,
    recipient_id uuid NOT NULL,
    inbox_revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    published_at timestamptz
);

-- +goose Down
DROP TABLE notifications_realtime_hints;
