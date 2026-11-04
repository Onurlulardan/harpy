-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_restored_license_scope_mint_v1(
    in_operation_id uuid,
    in_environment_id uuid,
    in_restore_receipt_id uuid,
    in_restore_receipt_sha256 text,
    in_license_id uuid,
    in_licensee_organization_id uuid,
    in_delivery_kind text,
    in_issuer text,
    in_previous_sequence bigint,
    in_previous_sha256 text,
    in_next_sequence bigint,
    in_next_sha256 text,
    in_manifest_sha256 text,
    in_authority_revision bigint
)
RETURNS TABLE (
    operation_id uuid,
    environment_id uuid,
    restore_receipt_id uuid,
    restore_receipt_sha256 text,
    license_binding_id uuid,
    license_id uuid,
    licensee_organization_id uuid,
    delivery_kind text,
    issuer text,
    previous_sequence bigint,
    previous_sha256 text,
    next_sequence bigint,
    next_sha256 text,
    manifest_sha256 text,
    revision bigint,
    lease_until timestamptz,
    database_now timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        0::bigint,
        ''::text,
        0::bigint,
        ''::text,
        ''::text,
        0::bigint,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp();
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_restored_license_scope_mint_v1(uuid, uuid, uuid, text, uuid, uuid, text, text, bigint, text, bigint, text, text, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_restored_license_scope_mint_v1(uuid, uuid, uuid, text, uuid, uuid, text, text, bigint, text, bigint, text, text, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_restored_license_scope_mint_v1(uuid, uuid, uuid, text, uuid, uuid, text, text, bigint, text, bigint, text, text, bigint);
