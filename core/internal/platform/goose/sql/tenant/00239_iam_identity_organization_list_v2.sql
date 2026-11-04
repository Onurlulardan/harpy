-- +goose Up
CREATE FUNCTION iam_identity_organization_list_v2(
    in_schema_name name,
    in_after_id uuid,
    in_limit integer,
    in_actor_id uuid,
    in_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    id uuid,
    name text,
    slug text,
    status text,
    parent_id uuid,
    revision bigint,
    created_at timestamptz,
    updated_at timestamptz,
    tombstoned_at timestamptz,
    owner_ids uuid[]
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        ''::text,
        NULL::uuid,
        0::bigint,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        NULL::timestamptz,
        '{}'::uuid[]
    ;
$$;

REVOKE ALL ON FUNCTION iam_identity_organization_list_v2(name, uuid, integer, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_identity_organization_list_v2(name, uuid, integer, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_identity_organization_list_v2(name, uuid, integer, uuid, uuid, bigint, bigint);
