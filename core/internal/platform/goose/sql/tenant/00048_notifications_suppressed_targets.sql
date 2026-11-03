-- +goose Up
CREATE TABLE notifications_suppressed_targets (
    channel text NOT NULL PRIMARY KEY,
    target_digest text NOT NULL,
    reason text NOT NULL,
    created_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE notifications_suppressed_targets;
