-- +goose Up
CREATE TABLE notifications_preferences (
    recipient_id uuid NOT NULL PRIMARY KEY,
    category text NOT NULL,
    channel text NOT NULL,
    enabled boolean NOT NULL,
    explicit text NOT NULL,
    verified_target text NOT NULL,
    quiet_timezone text NOT NULL,
    quiet_start_minute text NOT NULL,
    quiet_end_minute text NOT NULL,
    row_version bigint NOT NULL,
    updated_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE notifications_preferences;
