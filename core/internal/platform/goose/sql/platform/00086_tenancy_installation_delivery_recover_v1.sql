-- +goose Up
CREATE FUNCTION platform.tenancy_installation_delivery_recover_v1(
    in_operation_id uuid,
    in_current_attempt_id uuid,
    in_current_credential_id uuid,
    in_current_credential_version bigint,
    in_next_attempt_id uuid,
    in_next_credential_id uuid,
    in_next_credential_version bigint,
    in_expected_revision bigint,
    in_recovery_token_sha256 text,
    in_delivery_manifest_sha256 text,
    in_lease_seconds integer
)
RETURNS TABLE (
    operation_id uuid,
    tenant_id uuid,
    first_owner_user_id uuid,
    current_credential_id uuid,
    current_credential_version bigint,
    next_credential_id uuid,
    next_credential_version bigint,
    current_attempt_id uuid,
    next_attempt_id uuid,
    recovery_cause text,
    manifest_sha256 text,
    delivery_revision bigint,
    database_now timestamptz,
    lease_until timestamptz
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
        in_current_credential_id,
        in_current_credential_version,
        in_next_credential_id,
        in_next_credential_version,
        in_current_attempt_id,
        in_next_attempt_id,
        ''::text,
        in_delivery_manifest_sha256,
        in_expected_revision,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION platform.tenancy_installation_delivery_recover_v1(uuid, uuid, uuid, bigint, uuid, uuid, bigint, bigint, text, text, integer) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_installation_delivery_recover_v1(uuid, uuid, uuid, bigint, uuid, uuid, bigint, bigint, text, text, integer) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_installation_delivery_recover_v1(uuid, uuid, uuid, bigint, uuid, uuid, bigint, bigint, text, text, integer);
