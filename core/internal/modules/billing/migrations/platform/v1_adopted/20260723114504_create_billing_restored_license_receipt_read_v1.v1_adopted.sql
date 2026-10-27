-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_restored_license_receipt_read_v1(
    p_operation_id uuid,
    p_environment_id uuid
)
RETURNS TABLE (
    found boolean, operation_id uuid, restore_receipt_id uuid,
    restore_receipt_sha256 text, license_binding_id uuid,
    license_id uuid, licensee_organization_id uuid, environment_id uuid,
    previous_sequence bigint, previous_sha256 text,
    accepted_sequence bigint, accepted_sha256 text,
    authority_revision bigint, database_now timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_receipt platform.billing_license_restore_receipts%ROWTYPE;
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR pg_catalog.current_setting('application_name', true)
            <> 'forgecrud-v2:core-operator-license-reconcile'
       OR p_operation_id IS NULL
       OR pg_catalog.uuid_extract_version(p_operation_id) <> 7
       OR p_environment_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing restored license receipt read denied';
    END IF;
    IF NOT EXISTS (
        SELECT 1
          FROM platform.tenancy_deployment_policy AS policy
          JOIN platform.tenancy_tenant_bindings AS binding
            ON binding.environment_id = policy.environment_id
           AND binding.purpose = 'onpremise_primary'
         WHERE policy.singleton
           AND policy.environment_id = p_environment_id
           AND policy.deployment_mode = 'onpremise'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing restored license receipt requires onpremise primary';
    END IF;
    SELECT receipt.* INTO v_receipt
      FROM platform.billing_license_restore_receipts AS receipt
     WHERE receipt.operation_id = p_operation_id;
    IF NOT FOUND THEN
        RETURN QUERY SELECT false,
            '00000000-0000-0000-0000-000000000000'::uuid,
            '00000000-0000-0000-0000-000000000000'::uuid,
            ''::text, '00000000-0000-0000-0000-000000000000'::uuid,
            '00000000-0000-0000-0000-000000000000'::uuid,
            '00000000-0000-0000-0000-000000000000'::uuid,
            p_environment_id, 0::bigint, ''::text,
            0::bigint, ''::text, 0::bigint, v_now;
        RETURN;
    END IF;
    IF v_receipt.environment_id IS DISTINCT FROM p_environment_id THEN
        RAISE EXCEPTION USING ERRCODE = '23505',
            MESSAGE = 'restored license operation environment changed';
    END IF;
    RETURN QUERY SELECT true, v_receipt.operation_id,
        v_receipt.restore_receipt_id, v_receipt.restore_receipt_sha256,
        v_receipt.license_binding_id, v_receipt.license_id,
        v_receipt.licensee_organization_id, v_receipt.environment_id,
        v_receipt.previous_sequence, v_receipt.previous_sha256,
        v_receipt.accepted_sequence, v_receipt.accepted_sha256,
        v_receipt.authority_revision, v_now;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
platform.billing_restored_license_receipt_read_v1(uuid, uuid)
FROM PUBLIC;
