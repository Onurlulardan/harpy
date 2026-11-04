-- +goose Up
-- BELIRSIZ: slug lookup has no stored tenant. The id is the zero uuid.
CREATE FUNCTION platform.tenancy_runtime_resolve_slug_v1(
    p_slug text
)
RETURNS TABLE (
    id uuid,
    status text,
    not_deleted boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT '00000000-0000-0000-0000-000000000000'::uuid, ''::text, false;
$$;

REVOKE ALL ON FUNCTION platform.tenancy_runtime_resolve_slug_v1(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_runtime_resolve_slug_v1(text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_runtime_resolve_slug_v1(text);
