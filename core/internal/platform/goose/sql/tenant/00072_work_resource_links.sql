-- +goose Up
CREATE TABLE work_resource_links (
    id uuid NOT NULL PRIMARY KEY,
    work_resource_id uuid NOT NULL,
    target_owner text NOT NULL,
    target_kind text NOT NULL,
    target_id uuid NOT NULL,
    target_version_id uuid NOT NULL,
    target_reference_epoch text NOT NULL,
    permission_revision bigint NOT NULL,
    row_version bigint NOT NULL,
    work_resource_kind text NOT NULL,
    state text NOT NULL,
    usage text NOT NULL,
    created_at timestamptz NOT NULL,
    released_at timestamptz
);

-- +goose Down
DROP TABLE work_resource_links;
