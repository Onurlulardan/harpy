-- +goose Up
GRANT SELECT (schema_name, slug, status, deleted_at)
ON TABLE platform.tenants
TO forgecrud_runtime;
