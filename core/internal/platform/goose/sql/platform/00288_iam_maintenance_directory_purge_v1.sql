-- +goose Up
CREATE FUNCTION platform.iam_maintenance_directory_purge_v1(
    in_tenant_id uuid,
    in_identifier_id uuid,
    in_revision bigint
)
RETURNS bigint
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT 0::bigint;
$$;

REVOKE ALL ON FUNCTION platform.iam_maintenance_directory_purge_v1(uuid, uuid, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_maintenance_directory_purge_v1(uuid, uuid, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_maintenance_directory_purge_v1(uuid, uuid, bigint);
