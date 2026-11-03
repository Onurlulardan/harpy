-- +goose Up
CREATE TABLE documents_links (
    id uuid NOT NULL PRIMARY KEY,
    item_id uuid NOT NULL,
    lifecycle_state text NOT NULL,
    source_owner text NOT NULL,
    source_kind text NOT NULL,
    source_id uuid NOT NULL,
    source_version_id uuid NOT NULL,
    usage text NOT NULL,
    reference_epoch text NOT NULL,
    created_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE documents_links;
