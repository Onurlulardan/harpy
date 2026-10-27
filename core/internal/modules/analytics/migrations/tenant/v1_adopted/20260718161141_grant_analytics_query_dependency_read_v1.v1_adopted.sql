-- +goose Up
GRANT EXECUTE ON FUNCTION analytics_query_dependencies_v1(name,uuid,uuid)
    TO harpy_runtime;
