-- +goose Up
GRANT EXECUTE ON FUNCTION analytics_resource_manifest_v1(name,text,uuid)
    TO forgecrud_runtime;
