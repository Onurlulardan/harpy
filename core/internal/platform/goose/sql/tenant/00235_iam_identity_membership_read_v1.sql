-- +goose Up
CREATE FUNCTION iam_identity_membership_read_v1(
    in_schema_name name,
    in_id uuid,
    in_actor_id uuid,
    in_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    id uuid,
    organization_id uuid,
    user_id uuid,
    kind text,
    state text,
    revision bigint,
    created_at timestamptz,
    updated_at timestamptz,
    tombstoned_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        0::bigint,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        NULL::timestamptz
    ;
$$;

REVOKE ALL ON FUNCTION iam_identity_membership_read_v1(name, uuid, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_identity_membership_read_v1(name, uuid, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_identity_membership_read_v1(name, uuid, uuid, uuid, bigint, bigint);
