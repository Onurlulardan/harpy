-- +goose Up
CREATE TABLE notifications_provider_secrets (
    id uuid NOT NULL PRIMARY KEY,
    ciphertext text NOT NULL,
    nonce text NOT NULL,
    wrapped_dek text NOT NULL,
    kek_revision bigint NOT NULL,
    plaintext_digest text NOT NULL,
    created_at timestamptz NOT NULL,
    rewrapped_at timestamptz
);

-- +goose Down
DROP TABLE notifications_provider_secrets;
