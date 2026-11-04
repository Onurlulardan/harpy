-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_password_use_v1(
    in_schema_name name,
    in_user_id uuid,
    in_expected_user_revision bigint,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint,
    in_expected_credential_version bigint,
    in_expected_credential_revision bigint,
    in_rehash boolean,
    in_nullable_bytes bytea,
    in_nullable_bytes_2 bytea,
    in_memory integer,
    in_iterations smallint,
    in_parallelism smallint
)
RETURNS TABLE (
    user_revision bigint,
    token_version bigint,
    authorization_version bigint,
    credential_version bigint,
    credential_revision bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_password_use_v1(name, uuid, bigint, bigint, bigint, bigint, bigint, boolean, bytea, bytea, integer, smallint, smallint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_password_use_v1(name, uuid, bigint, bigint, bigint, bigint, bigint, boolean, bytea, bytea, integer, smallint, smallint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_password_use_v1(name, uuid, bigint, bigint, bigint, bigint, bigint, boolean, bytea, bytea, integer, smallint, smallint);
