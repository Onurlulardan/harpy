-- +goose Up
GRANT EXECUTE ON FUNCTION iam_lifecycle_credential_lookup_v1(
    name,text,bytea,bytea
) TO forgecrud_runtime;
