-- +goose Up
-- BELIRSIZ: receipt lookup is not determined by the caller.
CREATE FUNCTION platform.billing_restored_license_receipt_read_v1(
    p_operation_id uuid,
    p_environment_id uuid
)
RETURNS TABLE (
    found boolean,
    operation_id uuid,
    restore_receipt_id uuid,
    restore_receipt_sha256 text,
    license_binding_id uuid,
    license_id uuid,
    licensee_organization_id uuid,
    environment_id uuid,
    previous_sequence bigint,
    previous_sha256 text,
    accepted_sequence bigint,
    accepted_sha256 text,
    authority_revision bigint,
    database_now timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        false,
        p_operation_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        p_environment_id,
        0::bigint,
        ''::text,
        0::bigint,
        ''::text,
        0::bigint,
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION platform.billing_restored_license_receipt_read_v1(
    uuid, uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_restored_license_receipt_read_v1(
    uuid, uuid
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_restored_license_receipt_read_v1(uuid, uuid);
