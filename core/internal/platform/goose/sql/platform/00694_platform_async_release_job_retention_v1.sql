-- +goose Up
CREATE FUNCTION platform.platform_async_release_job_retention_v1(
    in_job_id uuid,
    in_scope_id uuid,
    in_group_id text,
    in_group_generation bigint,
    in_domain text,
    in_generation bigint,
    in_arg_7 bigint,
    in_digest text,
    in_retain_until timestamptz
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.platform_async_release_job_retention_v1(uuid, uuid, text, bigint, text, bigint, bigint, text, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platform_async_release_job_retention_v1(uuid, uuid, text, bigint, text, bigint, bigint, text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platform_async_release_job_retention_v1(uuid, uuid, text, bigint, text, bigint, bigint, text, timestamptz);
