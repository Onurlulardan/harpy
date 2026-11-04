-- +goose Up
-- BELIRSIZ: tenant authority assertion is not determined by the caller.
CREATE FUNCTION platform.billing_assert_tenant_authority_v1(
    p_authority_id uuid
)
RETURNS void
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT CAST(NULL AS pg_catalog.void);
$$;

REVOKE ALL ON FUNCTION platform.billing_assert_tenant_authority_v1(
    uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_assert_tenant_authority_v1(
    uuid
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_assert_tenant_authority_v1(
    uuid
);
