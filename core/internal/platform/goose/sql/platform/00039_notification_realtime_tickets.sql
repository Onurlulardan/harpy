-- +goose Up
CREATE TABLE platform.notification_realtime_tickets (
    id uuid NOT NULL PRIMARY KEY,
    tenant_id uuid NOT NULL,
    user_id uuid NOT NULL,
    token_hash text NOT NULL,
    authorization_version bigint NOT NULL,
    origin text NOT NULL,
    created_at timestamptz NOT NULL,
    expires_at timestamptz,
    consumed_at timestamptz
);

-- +goose Down
DROP TABLE platform.notification_realtime_tickets;
