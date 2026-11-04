-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_self_profile_read_v1(
    in_schema_name name,
    in_actor_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    result_id uuid,
    email text,
    status text,
    revision bigint,
    created_at timestamptz,
    updated_at timestamptz,
    roles bytea,
    organizations bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        0::bigint,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        ''::bytea,
        ''::bytea
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_self_profile_read_v1(name, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_self_profile_read_v1(name, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_self_profile_read_v1(name, uuid, bigint, bigint);
