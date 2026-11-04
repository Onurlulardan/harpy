-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_recovery_codes_list_v1(
    in_schema_name name,
    in_user_id uuid,
    in_actor_id uuid,
    in_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    authenticator_id uuid,
    authenticator_revision bigint,
    code_id uuid,
    position smallint,
    state text,
    created_at timestamptz,
    consumed_at timestamptz,
    revoked_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::smallint,
        ''::text,
        pg_catalog.clock_timestamp(),
        NULL::timestamptz,
        NULL::timestamptz
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_recovery_codes_list_v1(name, uuid, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_recovery_codes_list_v1(name, uuid, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_recovery_codes_list_v1(name, uuid, uuid, uuid, bigint, bigint);
