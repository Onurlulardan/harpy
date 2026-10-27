-- +goose Up
ALTER TABLE iam_profile_field_revisions
    ADD COLUMN field_key text COLLATE "C",
    ADD COLUMN value_kind text,
    ADD COLUMN cardinality text,
    ADD COLUMN required boolean,
    ADD COLUMN classification text,
    ADD COLUMN normalization_profile text,
    ADD COLUMN field_state text,
    ADD COLUMN semantic_sha256 bytea;
