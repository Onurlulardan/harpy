-- +goose Up
CREATE TABLE notifications_delivery_attempts (
    delivery_id uuid NOT NULL,
    id uuid NOT NULL PRIMARY KEY,
    attempt_number bigint NOT NULL,
    lease_generation bigint NOT NULL,
    provider_kind text NOT NULL,
    response_class text NOT NULL,
    provider_reference_digest text NOT NULL,
    status_code text NOT NULL,
    error_code text NOT NULL,
    started_at timestamptz,
    finished_at timestamptz
);

-- +goose Down
DROP TABLE notifications_delivery_attempts;
