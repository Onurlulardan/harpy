-- +goose Up
-- BELIRSIZ: SaaS free base offering is not determined by the caller.
CREATE FUNCTION platform.billing_saas_free_base_read_v1()
RETURNS TABLE (
    offering_version_id uuid,
    revision bigint,
    offering_sha256 text,
    effective_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        ''::text,
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION platform.billing_saas_free_base_read_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_saas_free_base_read_v1() TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_saas_free_base_read_v1();
