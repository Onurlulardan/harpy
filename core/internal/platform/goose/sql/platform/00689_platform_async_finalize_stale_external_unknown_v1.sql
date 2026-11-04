-- +goose Up
CREATE FUNCTION platform.platform_async_finalize_stale_external_unknown_v1(
    in_arg_1 uuid,
    in_arg_2 uuid,
    in_arg_3 bigint,
    in_arg_4 uuid,
    in_arg_5 timestamptz
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.platform_async_finalize_stale_external_unknown_v1(uuid, uuid, bigint, uuid, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platform_async_finalize_stale_external_unknown_v1(uuid, uuid, bigint, uuid, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platform_async_finalize_stale_external_unknown_v1(uuid, uuid, bigint, uuid, timestamptz);
