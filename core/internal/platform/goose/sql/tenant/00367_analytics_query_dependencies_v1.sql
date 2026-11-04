-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION analytics_query_dependencies_v1(
    arg_1 name,
    arg_2 uuid,
    arg_3 uuid
)
RETURNS TABLE (
    owner_domain text,
    resource_kind text,
    resource_id uuid,
    resource_version_id uuid,
    owner_generation bigint,
    catalog_version bigint,
    query_contract_version bigint,
    policy_version bigint,
    data_revision_capability text,
    dependency_sha256 bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        ''::text,
        NULL::bytea;
$$;

REVOKE ALL ON FUNCTION analytics_query_dependencies_v1(name, uuid, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_query_dependencies_v1(name, uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_query_dependencies_v1(name, uuid, uuid);
