-- +goose Up
CREATE TABLE platform.billing_webhook_receipts (
    id uuid NOT NULL PRIMARY KEY,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    provider_account_id uuid NOT NULL,
    provider_object_mapping_id uuid NOT NULL,
    provider_event_id uuid NOT NULL,
    body_sha256 text NOT NULL,
    authentication_digest text NOT NULL,
    verification_revision bigint NOT NULL,
    event_occurred_at timestamptz,
    signature_issued_at timestamptz,
    algorithm text NOT NULL,
    nonce text NOT NULL,
    encrypted_body text NOT NULL,
    wrapped_dek text NOT NULL,
    key_id uuid NOT NULL,
    verified_at timestamptz,
    received_at timestamptz,
    processor_state text NOT NULL,
    processed_at timestamptz,
    raw_body_expires_at timestamptz,
    retain_until text NOT NULL,
    revision bigint NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_webhook_receipts;
