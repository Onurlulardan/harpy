-- +goose Up
-- BELIRSIZ: release has no stored lease to clear.
CREATE FUNCTION platform.analytics_public_release_v1(
    p_lease_id uuid,
    p_tenant_id uuid,
    p_release_id uuid,
    p_generation bigint
)
RETURNS void
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT CAST(NULL AS pg_catalog.void);
$$;

REVOKE ALL ON FUNCTION platform.analytics_public_release_v1(uuid, uuid, uuid, bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.analytics_public_release_v1(uuid, uuid, uuid, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.analytics_public_release_v1(uuid, uuid, uuid, bigint);
