-- +goose Up
CREATE TABLE analytics_query_dependencies (
    query_version_id uuid NOT NULL PRIMARY KEY
);

-- +goose Down
DROP TABLE analytics_query_dependencies;
