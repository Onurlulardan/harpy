-- +goose Up
CREATE TABLE documents_upload_attempts (
    upload_id uuid NOT NULL PRIMARY KEY,
    attempt_generation bigint NOT NULL,
    lifecycle_state text NOT NULL,
    object_key text NOT NULL,
    object_version bigint NOT NULL,
    object_generation bigint NOT NULL,
    plaintext_bytes bigint NOT NULL,
    ciphertext_bytes bigint NOT NULL,
    plaintext_sha256 text NOT NULL,
    ciphertext_sha256 text NOT NULL,
    effect_receipt_id uuid NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE documents_upload_attempts;
