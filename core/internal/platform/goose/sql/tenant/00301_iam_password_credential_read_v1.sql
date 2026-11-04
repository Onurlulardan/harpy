-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_password_credential_read_v1(
    in_schema_name name,
    in_user_id uuid,
    in_actor_id uuid,
    in_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    algorithm text,
    salt bytea,
    tag bytea,
    memory_ki_b bigint,
    iterations bigint,
    parallelism smallint,
    legacy_bcrypt_hash text,
    version bigint,
    revision bigint,
    created_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::text,
        ''::bytea,
        ''::bytea,
        0::bigint,
        0::bigint,
        0::smallint,
        NULL::text,
        0::bigint,
        0::bigint,
        pg_catalog.clock_timestamp()
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_password_credential_read_v1(name, uuid, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_password_credential_read_v1(name, uuid, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_password_credential_read_v1(name, uuid, uuid, uuid, bigint, bigint);
