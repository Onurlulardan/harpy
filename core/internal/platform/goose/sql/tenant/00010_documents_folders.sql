-- +goose Up
CREATE TABLE documents_folders (
    id uuid NOT NULL PRIMARY KEY,
    lifecycle_state text NOT NULL,
    parent_id uuid NOT NULL,
    owner_id uuid NOT NULL,
    display_name text NOT NULL,
    row_version bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    trashed_at timestamptz
);

-- +goose Down
DROP TABLE documents_folders;
