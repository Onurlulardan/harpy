-- +goose Up
CREATE FUNCTION iam_initial_password_start_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_user_id uuid,
    in_installation_operation_id uuid,
    in_credential_id uuid,
    in_credential_version bigint,
    in_credential_revision bigint,
    in_selector bytea,
    in_secret_key_id text,
    in_secret_hmac bytea,
    in_csrf_key_id text,
    in_csrf_hmac bytea
)
RETURNS TABLE (
    revision bigint,
    issued_at timestamptz,
    expires_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        0::bigint,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp()
    ;
$$;

REVOKE ALL ON FUNCTION iam_initial_password_start_v1(name, uuid, uuid, uuid, uuid, bigint, bigint, bytea, text, bytea, text, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_initial_password_start_v1(name, uuid, uuid, uuid, uuid, bigint, bigint, bytea, text, bytea, text, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_initial_password_start_v1(name, uuid, uuid, uuid, uuid, bigint, bigint, bytea, text, bytea, text, bytea);
