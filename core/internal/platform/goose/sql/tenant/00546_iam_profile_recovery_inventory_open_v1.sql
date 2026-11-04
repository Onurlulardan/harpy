-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_profile_recovery_inventory_open_v1(
    in_schema_name name,
    in_parent_id uuid,
    in_actor_id uuid,
    in_manage_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
)
RETURNS text
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT ''::text;
$$;

REVOKE ALL ON FUNCTION iam_profile_recovery_inventory_open_v1(name, uuid, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_profile_recovery_inventory_open_v1(name, uuid, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_profile_recovery_inventory_open_v1(name, uuid, uuid, uuid, bigint, bigint);
