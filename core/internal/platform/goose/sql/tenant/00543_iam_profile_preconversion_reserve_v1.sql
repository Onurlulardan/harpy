-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_profile_preconversion_reserve_v1(
    in_schema_name name,
    in_plan_id uuid,
    in_resource text,
    in_resource_id uuid,
    in_actual_schema_id uuid,
    in_revision bigint,
    in_source_sha bytea,
    in_planned_sha bytea,
    in_target_sha bytea
)
RETURNS bigint
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT 0::bigint;
$$;

REVOKE ALL ON FUNCTION iam_profile_preconversion_reserve_v1(name, uuid, text, uuid, uuid, bigint, bytea, bytea, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_profile_preconversion_reserve_v1(name, uuid, text, uuid, uuid, bigint, bytea, bytea, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_profile_preconversion_reserve_v1(name, uuid, text, uuid, uuid, bigint, bytea, bytea, bytea);
