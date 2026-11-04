-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_initial_password_continuation_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_selector bytea
)
RETURNS TABLE (
    "or" bytea,
    tenant_id uuid,
    user_id uuid,
    installation_operation_id uuid,
    credential_id uuid,
    credential_version bigint,
    credential_revision bigint,
    secret_key_id text,
    secret_hmac bytea,
    csrf_key_id text,
    csrf_hmac bytea,
    state text,
    policy_attempts smallint,
    revision bigint,
    issued_at timestamptz,
    expires_at timestamptz,
    updated_at timestamptz,
    consumed_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::bytea,
        in_tenant_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        0::bigint,
        ''::text,
        ''::bytea,
        ''::text,
        ''::bytea,
        ''::text,
        0::smallint,
        0::bigint,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        NULL::timestamptz
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_initial_password_continuation_v1(name, uuid, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_initial_password_continuation_v1(name, uuid, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_initial_password_continuation_v1(name, uuid, bytea);
