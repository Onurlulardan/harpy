-- +goose Up
CREATE TABLE notifications_provider_configs (
    id uuid NOT NULL PRIMARY KEY,
    provider_kind text NOT NULL,
    name text NOT NULL,
    enabled boolean NOT NULL,
    current_version_id uuid NOT NULL,
    row_version bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE notifications_provider_configs;
