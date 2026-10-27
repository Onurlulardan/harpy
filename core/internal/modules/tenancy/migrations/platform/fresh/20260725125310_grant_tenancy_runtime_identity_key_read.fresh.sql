-- +goose Up
GRANT SELECT (id)
ON TABLE platform.tenants
TO harpy_runtime;
