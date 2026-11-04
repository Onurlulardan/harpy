-- +goose Up
CREATE FUNCTION platform.platform_async_event_gap_job_ready_v2(
    in_job_id uuid
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.platform_async_event_gap_job_ready_v2(uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platform_async_event_gap_job_ready_v2(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platform_async_event_gap_job_ready_v2(uuid);
