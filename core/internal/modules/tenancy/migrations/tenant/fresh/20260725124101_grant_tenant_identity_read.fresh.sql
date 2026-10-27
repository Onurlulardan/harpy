-- +goose Up
GRANT SELECT ON TABLE tenancy_schema_identity TO harpy_runtime;

