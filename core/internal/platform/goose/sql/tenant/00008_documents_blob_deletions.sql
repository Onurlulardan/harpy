-- +goose Up
CREATE TABLE documents_blob_deletions (
    id uuid NOT NULL PRIMARY KEY,
    target_kind text NOT NULL,
    target_id uuid NOT NULL,
    object_key text NOT NULL,
    object_version bigint NOT NULL,
    object_generation bigint NOT NULL,
    plaintext_bytes bigint NOT NULL,
    ciphertext_bytes bigint NOT NULL,
    lifecycle_state text NOT NULL,
    lease_generation bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    confirmed_at timestamptz,
    lease_owner_id uuid NOT NULL,
    lease_expires_at timestamptz
);

-- +goose Down
DROP TABLE documents_blob_deletions;
