-- +goose Up
CREATE TABLE documents_archives (
    lifecycle_state text NOT NULL,
    ready_at timestamptz,
    expires_at timestamptz,
    updated_at timestamptz NOT NULL,
    id uuid NOT NULL PRIMARY KEY,
    entry_count bigint NOT NULL,
    object_version bigint NOT NULL,
    artifact_plaintext_bytes bigint NOT NULL,
    ciphertext_bytes bigint NOT NULL,
    artifact_sha256 text NOT NULL,
    ciphertext_sha256 text NOT NULL,
    cipher_format text NOT NULL,
    object_key text NOT NULL,
    object_generation bigint NOT NULL,
    key_id uuid NOT NULL,
    wrapped_dek text NOT NULL,
    actor_id uuid NOT NULL,
    authorization_revision bigint NOT NULL,
    job_id uuid NOT NULL,
    retry_of_archive_id uuid NOT NULL,
    request_sha256 text NOT NULL,
    request_document text NOT NULL,
    billing_reservation_id uuid NOT NULL,
    plaintext_bytes bigint NOT NULL,
    created_at timestamptz NOT NULL,
    upload_id uuid NOT NULL,
    attempt_generation bigint NOT NULL
);

-- +goose Down
DROP TABLE documents_archives;
