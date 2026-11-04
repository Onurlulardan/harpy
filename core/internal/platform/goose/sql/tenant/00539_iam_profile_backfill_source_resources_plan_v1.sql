-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_profile_backfill_source_resources_plan_v1(
    in_schema_name name,
    in_resource_kind text,
    in_source_schema_id uuid,
    in_after_resource_id uuid,
    in_limit integer,
    in_actor_id uuid,
    in_manage_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    resource_id uuid,
    source_schema_version_id uuid,
    source_revision bigint,
    profile_values bytea,
    profile_sha256 bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        ''::bytea,
        ''::bytea
    ;
$$;

REVOKE ALL ON FUNCTION iam_profile_backfill_source_resources_plan_v1(name, text, uuid, uuid, integer, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_profile_backfill_source_resources_plan_v1(name, text, uuid, uuid, integer, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_profile_backfill_source_resources_plan_v1(name, text, uuid, uuid, integer, uuid, uuid, bigint, bigint);
