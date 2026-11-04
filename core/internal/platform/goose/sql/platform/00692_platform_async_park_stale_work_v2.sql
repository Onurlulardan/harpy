-- +goose Up
CREATE FUNCTION platform.platform_async_park_stale_work_v2(
    in_job_id uuid,
    in_state text,
    in_lease_owner_argument uuid,
    in_lease_generation bigint,
    in_new uuid
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.platform_async_park_stale_work_v2(uuid, text, uuid, bigint, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platform_async_park_stale_work_v2(uuid, text, uuid, bigint, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platform_async_park_stale_work_v2(uuid, text, uuid, bigint, uuid);
