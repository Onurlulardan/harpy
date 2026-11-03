-- +goose Up
CREATE TABLE notifications_effect_receipts (
    message_id uuid NOT NULL PRIMARY KEY,
    outcome text NOT NULL,
    producer text NOT NULL,
    effect_key_hash text NOT NULL,
    recipient_id uuid NOT NULL,
    canonical_digest text NOT NULL,
    legacy_receipt_reference text NOT NULL,
    created_at timestamptz NOT NULL,
    retain_until text NOT NULL
);

-- +goose Down
DROP TABLE notifications_effect_receipts;
