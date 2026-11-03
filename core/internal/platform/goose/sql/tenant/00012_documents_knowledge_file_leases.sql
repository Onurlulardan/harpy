-- +goose Up
CREATE TABLE documents_knowledge_file_leases (
    reference_id uuid NOT NULL PRIMARY KEY,
    source_id uuid NOT NULL,
    lifecycle_state text NOT NULL,
    adopted_at timestamptz,
    updated_at timestamptz NOT NULL,
    lease_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    item_id uuid NOT NULL,
    version_id uuid NOT NULL,
    upload_id uuid NOT NULL,
    expires_at timestamptz,
    expected_plaintext_bytes bigint NOT NULL,
    expected_plaintext_sha256 text NOT NULL,
    filename text NOT NULL,
    declared_mime text NOT NULL,
    detected_mime text NOT NULL,
    scan_signature_revision bigint NOT NULL,
    scanned_at timestamptz,
    released_at timestamptz,
    created_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE documents_knowledge_file_leases;
