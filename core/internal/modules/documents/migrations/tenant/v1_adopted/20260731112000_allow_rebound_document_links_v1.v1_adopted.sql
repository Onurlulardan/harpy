-- +goose Up
-- +goose StatementBegin
ALTER TABLE documents_links
    DROP CONSTRAINT documents_links_item_id_source_owner_source_kind_source_id__key;

CREATE UNIQUE INDEX documents_links_active_source_target_unique
    ON documents_links (
        item_id,
        source_owner,
        source_kind,
        source_id,
        source_version_id,
        usage
    )
    WHERE lifecycle_state = 'ACTIVE';
-- +goose StatementEnd
