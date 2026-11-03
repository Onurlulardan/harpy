-- +goose Up
CREATE TABLE notifications_provider_config_versions (
    id uuid NOT NULL PRIMARY KEY,
    config_id uuid NOT NULL,
    version_number bigint NOT NULL,
    secret_id uuid NOT NULL,
    metadata jsonb NOT NULL,
    metadata_digest text NOT NULL,
    created_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE notifications_provider_config_versions;
