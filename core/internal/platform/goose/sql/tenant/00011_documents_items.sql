-- +goose Up
CREATE TABLE documents_items (
    id uuid NOT NULL PRIMARY KEY,
    folder_id uuid NOT NULL,
    owner_id uuid NOT NULL,
    current_version_id uuid NOT NULL,
    lifecycle_state text NOT NULL,
    row_version bigint NOT NULL,
    description text NOT NULL,
    tags jsonb NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    version_number bigint NOT NULL,
    purged_at timestamptz,
    purge_starts_at timestamptz,
    retention_ends_at timestamptz,
    reference_epoch text NOT NULL,
    trashed_at timestamptz,
    purge_requested_at timestamptz,
    resume_state text NOT NULL
);

-- +goose Down
DROP TABLE documents_items;
