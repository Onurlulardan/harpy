-- +goose Up
UPDATE iam_profile_schema_versions AS schema_version
SET definition_sha256 = iam_profile_definition_sha256_v1(schema_version.id);
