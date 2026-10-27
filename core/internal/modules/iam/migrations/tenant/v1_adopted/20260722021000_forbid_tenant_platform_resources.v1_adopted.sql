-- +goose Up
ALTER TABLE iam_resources
ADD CONSTRAINT iam_resources_platform_scope_ck
CHECK (resource_key !~ '^platform\.');
