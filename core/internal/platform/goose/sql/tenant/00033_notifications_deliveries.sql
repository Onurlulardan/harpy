-- +goose Up
CREATE TABLE notifications_deliveries (
    message_id uuid NOT NULL,
    id uuid NOT NULL PRIMARY KEY,
    provider_config_version_id uuid NOT NULL,
    channel text NOT NULL,
    state text NOT NULL,
    attempt text NOT NULL,
    maximum_attempts text NOT NULL,
    lease_owner text NOT NULL,
    lease_generation bigint NOT NULL,
    lease_until timestamptz,
    not_before text NOT NULL,
    outcome_code text NOT NULL,
    outcome_unknown text NOT NULL,
    mandatory text NOT NULL,
    row_version bigint NOT NULL
);

-- +goose Down
DROP TABLE notifications_deliveries;
