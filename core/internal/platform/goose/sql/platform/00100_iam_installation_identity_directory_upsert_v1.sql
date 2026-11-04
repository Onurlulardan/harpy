-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_installation_identity_directory_upsert_v1(
    in_operation_id uuid,
    in_environment_id uuid,
    in_tenant_id uuid,
    in_schema_name name,
    in_bound_bigint_5 bigint,
    in_bound_bigint_6 bigint,
    in_user_id uuid,
    in_identifier_id uuid,
    in_key_id text,
    in_identifier_hmac bytea,
    in_bound_bigint_11 bigint,
    in_source_sha256 bytea
)
RETURNS bigint
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT 0::bigint;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_installation_identity_directory_upsert_v1(uuid, uuid, uuid, name, bigint, bigint, uuid, uuid, text, bytea, bigint, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_installation_identity_directory_upsert_v1(uuid, uuid, uuid, name, bigint, bigint, uuid, uuid, text, bytea, bigint, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_installation_identity_directory_upsert_v1(uuid, uuid, uuid, name, bigint, bigint, uuid, uuid, text, bytea, bigint, bytea);
