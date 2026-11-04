-- +goose Up
CREATE FUNCTION platform.platform_async_assert_worker_authority_v1(
    in_scope_kind text,
    in_scope_id uuid,
    in_group_id text,
    in_group_generation bigint,
    in_domain text,
    in_cutover_generation bigint,
    in_mode text
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.platform_async_assert_worker_authority_v1(text, uuid, text, bigint, text, bigint, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platform_async_assert_worker_authority_v1(text, uuid, text, bigint, text, bigint, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platform_async_assert_worker_authority_v1(text, uuid, text, bigint, text, bigint, text);
