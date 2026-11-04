-- +goose Up
CREATE FUNCTION platform.platformops_prune_observability_v1(
    in_environment_id uuid,
    in_before timestamptz,
    in_limit integer
)
RETURNS bigint
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT 0::bigint;
$$;

REVOKE ALL ON FUNCTION platform.platformops_prune_observability_v1(uuid, timestamptz, integer) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_prune_observability_v1(uuid, timestamptz, integer) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_prune_observability_v1(uuid, timestamptz, integer);
