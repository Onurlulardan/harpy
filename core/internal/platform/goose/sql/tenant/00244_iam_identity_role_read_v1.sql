-- +goose Up
CREATE FUNCTION iam_identity_role_read_v1(
    in_schema_name name,
    in_id uuid,
    in_actor_id uuid,
    in_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    id uuid,
    scope_kind text,
    scope_id uuid,
    organization_id uuid,
    name text,
    description text,
    status text,
    protected boolean,
    "default" boolean,
    revision bigint,
    created_at timestamptz,
    updated_at timestamptz,
    retired_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_id,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        NULL::uuid,
        ''::text,
        ''::text,
        ''::text,
        false,
        false,
        0::bigint,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        NULL::timestamptz
    ;
$$;

REVOKE ALL ON FUNCTION iam_identity_role_read_v1(name, uuid, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_identity_role_read_v1(name, uuid, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_identity_role_read_v1(name, uuid, uuid, uuid, bigint, bigint);
