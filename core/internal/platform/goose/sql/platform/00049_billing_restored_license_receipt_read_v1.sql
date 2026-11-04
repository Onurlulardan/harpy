-- +goose Up
-- +goose StatementBegin
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
AS $fn$
    SELECT
        true, receipt.operation_id, receipt.restore_receipt_id, receipt.restore_receipt_sha256,
        receipt.license_binding_id, receipt.license_id, receipt.licensee_organization_id,
        receipt.environment_id, receipt.previous_sequence, receipt.previous_sha256,
        receipt.accepted_sequence, receipt.accepted_sha256, receipt.authority_revision,
        receipt.database_now
    FROM platform.billing_license_restore_receipts AS receipt
    WHERE receipt.operation_id = p_operation_id
      AND receipt.environment_id = p_environment_id
    UNION ALL
    SELECT
        false, p_operation_id,
        '00000000-0000-0000-0000-000000000000'::uuid, NULL::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        p_environment_id, 0, NULL::text, 0, NULL::text, 0,
        pg_catalog.clock_timestamp()
    WHERE NOT EXISTS (
        SELECT 1 FROM platform.billing_license_restore_receipts AS receipt
        WHERE receipt.operation_id = p_operation_id
          AND receipt.environment_id = p_environment_id
    );
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_restored_license_receipt_read_v1(uuid, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_restored_license_receipt_read_v1(uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_restored_license_receipt_read_v1(uuid, uuid);
