-- +goose Up
CREATE TABLE documents_versions (
    id uuid NOT NULL PRIMARY KEY,
    item_id uuid NOT NULL,
    version_number bigint NOT NULL,
    lifecycle_state text NOT NULL,
    original_name text NOT NULL,
    declared_mime text NOT NULL,
    detected_mime text NOT NULL,
    file_kind text NOT NULL,
    plaintext_bytes bigint NOT NULL,
    ciphertext_bytes bigint NOT NULL,
    plaintext_sha256 text NOT NULL,
    ciphertext_sha256 text NOT NULL,
    cipher_format text NOT NULL,
    object_key text NOT NULL,
    object_version bigint NOT NULL,
    object_generation bigint NOT NULL,
    key_id uuid NOT NULL,
    wrapped_dek text NOT NULL,
    scan_signature_revision bigint NOT NULL,
    scanned_at timestamptz,
    created_at timestamptz NOT NULL,
    delete_receipt_sha256 text NOT NULL
);

-- +goose Down
DROP TABLE documents_versions;
