-- +goose Up
CREATE TABLE notifications_inbox_counters (
    recipient_id uuid NOT NULL PRIMARY KEY,
    current_revision bigint NOT NULL
);

-- +goose Down
DROP TABLE notifications_inbox_counters;
