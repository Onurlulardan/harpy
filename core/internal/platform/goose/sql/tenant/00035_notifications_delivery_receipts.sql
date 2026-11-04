-- +goose Up
CREATE TABLE notifications_delivery_receipts (
    delivery_id uuid NOT NULL,
    id uuid NOT NULL PRIMARY KEY,
    provider_reference_digest bytea NOT NULL,
    outcome text NOT NULL,
    signature_key_id uuid NOT NULL,
    received_at timestamptz
);

-- +goose Down
DROP TABLE notifications_delivery_receipts;
