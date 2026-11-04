-- +goose Up
-- BELIRSIZ: stored restore binding columns are not determined by the caller.
CREATE FUNCTION platform.billing_license_restore_binding_v1(
    p_environment_id uuid
)
RETURNS TABLE (
    license_id uuid,
    licensee_organization_id uuid,
    environment_id uuid,
    delivery_kind text,
    issuer text,
    sequence bigint,
    payload_sha256 text,
    authority_revision bigint,
    database_now timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        p_environment_id,
        ''::text,
        ''::text,
        0::bigint,
        ''::text,
        0::bigint,
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION platform.billing_license_restore_binding_v1(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_license_restore_binding_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_license_restore_binding_v1(uuid);
