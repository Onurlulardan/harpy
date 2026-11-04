-- +goose Up
CREATE TABLE documents_archive_entries (
    archive_id uuid NOT NULL PRIMARY KEY,
    entry_order text NOT NULL,
    document_id uuid NOT NULL,
    version_id uuid NOT NULL,
    plaintext_bytes bigint NOT NULL,
    plaintext_sha256 text NOT NULL,
    added_at timestamptz
);

-- +goose Down
DROP TABLE documents_archive_entries;
