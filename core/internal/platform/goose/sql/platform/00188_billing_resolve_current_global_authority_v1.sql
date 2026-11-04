-- +goose Up
-- BELIRSIZ: current global authority is not determined by the caller.
CREATE FUNCTION platform.billing_resolve_current_global_authority_v1(
    p_domain text
)
RETURNS uuid
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT '00000000-0000-0000-0000-000000000000'::uuid;
$$;

REVOKE ALL ON FUNCTION platform.billing_resolve_current_global_authority_v1(
    text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_resolve_current_global_authority_v1(
    text
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_resolve_current_global_authority_v1(
    text
);
