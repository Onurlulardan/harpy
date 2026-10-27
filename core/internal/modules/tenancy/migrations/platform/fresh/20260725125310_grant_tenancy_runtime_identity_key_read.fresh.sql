-- +goose Up
GRANT SELECT (id)
ON TABLE platform.tenants
TO forgecrud_runtime;
