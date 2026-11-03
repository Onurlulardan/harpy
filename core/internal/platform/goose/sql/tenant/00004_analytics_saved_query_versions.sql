-- +goose Up
CREATE TABLE analytics_saved_query_versions (
    query_id uuid NOT NULL,
    id uuid NOT NULL PRIMARY KEY,
    manifest_sha256 text NOT NULL,
    query_document text NOT NULL,
    parameter_document text NOT NULL,
    output_document text NOT NULL,
    version_state text NOT NULL,
    row_revision bigint NOT NULL
);

-- +goose Down
DROP TABLE analytics_saved_query_versions;
