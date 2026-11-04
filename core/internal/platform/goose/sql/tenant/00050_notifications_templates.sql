-- +goose Up
CREATE TABLE notifications_templates (
    id uuid NOT NULL PRIMARY KEY,
    row_version bigint NOT NULL,
    notification_type text NOT NULL,
    published_version_id uuid NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE notifications_templates;
