-- +goose Up
CREATE FUNCTION platform.tenancy_installation_delivery_read_v1(
    in_operation_id uuid
)
RETURNS TABLE (
    operation_id uuid,
    tenant_id uuid,
    first_owner_user_id uuid,
    current_attempt_id uuid,
    current_credential_id uuid,
    current_credential_version bigint,
    delivery_state text,
    delivery_revision bigint,
    database_now timestamptz,
    claim_adapter_id text,
    claim_lease_until timestamptz,
    acknowledged_at timestamptz,
    secret_deleted_at timestamptz,
    completed_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_operation_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        ''::text,
        0::bigint,
        pg_catalog.clock_timestamp(),
        NULL::text,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::timestamptz;
$$;

REVOKE ALL ON FUNCTION platform.tenancy_installation_delivery_read_v1(uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_installation_delivery_read_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_installation_delivery_read_v1(uuid);
