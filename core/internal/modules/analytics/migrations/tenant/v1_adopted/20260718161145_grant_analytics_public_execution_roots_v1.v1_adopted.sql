-- +goose Up
GRANT EXECUTE ON FUNCTION analytics_public_query_root_v1(name,uuid,uuid)
    TO harpy_runtime;
GRANT EXECUTE ON FUNCTION analytics_public_dashboard_root_v1(name,uuid,uuid)
    TO harpy_runtime;
