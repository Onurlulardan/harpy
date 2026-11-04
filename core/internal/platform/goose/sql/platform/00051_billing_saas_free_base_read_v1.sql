-- +goose Up
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
    SELECT offering.offering_version_id, offering.revision, offering.offering_sha256,
           offering.effective_at
    FROM platform.billing_license_offerings AS offering
    WHERE offering.deployment_mode = 'saas'
    ORDER BY offering.revision DESC
    LIMIT 1;
$$;

REVOKE ALL ON FUNCTION platform.billing_saas_free_base_read_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_saas_free_base_read_v1() TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_saas_free_base_read_v1();
