-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_restored_license_scope_mint_v1(
    p_operation_id uuid, p_environment_id uuid, p_restore_receipt_id uuid,
    p_restore_receipt_sha256 text, p_license_id uuid,
    p_licensee_organization_id uuid, p_delivery_kind text, p_issuer text,
    p_previous_sequence bigint, p_previous_sha256 text,
    p_next_sequence bigint, p_next_sha256 text,
    p_manifest_sha256 text, p_authority_revision bigint
)
RETURNS TABLE (
    operation_id uuid, environment_id uuid, restore_receipt_id uuid,
    restore_receipt_sha256 text, license_binding_id uuid, license_id uuid,
    licensee_organization_id uuid, delivery_kind text, issuer text,
    previous_sequence bigint, previous_sha256 text,
    next_sequence bigint, next_sha256 text, manifest_sha256 text,
    authority_revision bigint, lease_until timestamptz, database_now timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.clock_timestamp();
    v_restore platform.platformops_restore_operations%ROWTYPE;
    v_grant platform.platformops_restored_license_grants%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR pg_catalog.current_setting('application_name', true)
            <> 'harpy-v2:core-operator-license-reconcile' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'restored license scope mint denied';
    END IF;
    IF pg_catalog.uuid_extract_version(p_operation_id) <> 7
       OR p_environment_id IS NULL OR p_restore_receipt_id IS NULL
       OR p_license_id IS NULL OR p_licensee_organization_id IS NULL
       OR p_restore_receipt_sha256 !~ '^[0-9a-f]{64}$'
       OR p_delivery_kind NOT IN ('connected','airgap','development')
       OR pg_catalog.octet_length(p_issuer) NOT BETWEEN 1 AND 256
       OR p_issuer <> pg_catalog.btrim(p_issuer) OR p_issuer ~ '[[:cntrl:]]'
       OR p_previous_sequence < 1 OR p_next_sequence <= p_previous_sequence
       OR p_previous_sha256 !~ '^[0-9a-f]{64}$'
       OR p_next_sha256 !~ '^[0-9a-f]{64}$'
       OR p_manifest_sha256 <>
          '982f8d8b6315f237e87a80904ceda1ed4ef54d32931907cda9be1db07a2f5d45'
       OR p_authority_revision < 1 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'restored license scope input is invalid';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM platform.tenancy_deployment_policy AS policy
        JOIN platform.tenancy_tenant_bindings AS binding
          ON binding.environment_id = policy.environment_id
         AND binding.purpose = 'onpremise_primary'
        WHERE policy.singleton AND policy.deployment_mode = 'onpremise'
          AND policy.environment_id = p_environment_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'restored license scope requires onpremise primary';
    END IF;
    SELECT restore.* INTO v_restore
      FROM platform.platformops_restore_operations AS restore
     WHERE restore.operation_id = p_operation_id
     FOR SHARE;
    IF NOT FOUND OR v_restore.state <> 'completed'
       OR v_restore.environment_id IS DISTINCT FROM p_environment_id
       OR v_restore.receipt_sha256 IS DISTINCT FROM p_restore_receipt_sha256
       OR v_restore.receipt_bytes IS NULL
       OR pg_catalog.encode(pg_catalog.sha256(v_restore.receipt_bytes), 'hex')
            IS DISTINCT FROM p_restore_receipt_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'completed restore receipt binding is invalid';
    END IF;
    INSERT INTO platform.platformops_restored_license_grants (
        operation_id, environment_id, restore_receipt_id,
        restore_receipt_sha256, license_binding_id, license_id,
        licensee_organization_id, delivery_kind, issuer,
        previous_sequence, previous_sha256, next_sequence, next_sha256,
        manifest_sha256, authority_revision, lease_until, created_at, updated_at
    ) VALUES (
        p_operation_id, p_environment_id, p_restore_receipt_id,
        p_restore_receipt_sha256, pg_catalog.uuidv7(), p_license_id,
        p_licensee_organization_id, p_delivery_kind, p_issuer,
        p_previous_sequence, p_previous_sha256, p_next_sequence, p_next_sha256,
        p_manifest_sha256, p_authority_revision, v_now + interval '60 seconds',
        v_now, v_now
    ) ON CONFLICT ON CONSTRAINT platformops_restored_license_grants_pkey
      DO NOTHING;
    SELECT grant_row.* INTO v_grant
      FROM platform.platformops_restored_license_grants AS grant_row
     WHERE grant_row.operation_id = p_operation_id FOR UPDATE;
    IF v_grant.environment_id IS DISTINCT FROM p_environment_id
       OR v_grant.restore_receipt_id IS DISTINCT FROM p_restore_receipt_id
       OR v_grant.restore_receipt_sha256 IS DISTINCT FROM p_restore_receipt_sha256
       OR v_grant.license_id IS DISTINCT FROM p_license_id
       OR v_grant.licensee_organization_id IS DISTINCT FROM p_licensee_organization_id
       OR v_grant.delivery_kind IS DISTINCT FROM p_delivery_kind
       OR v_grant.issuer IS DISTINCT FROM p_issuer
       OR v_grant.previous_sequence IS DISTINCT FROM p_previous_sequence
       OR v_grant.previous_sha256 IS DISTINCT FROM p_previous_sha256
       OR v_grant.next_sequence IS DISTINCT FROM p_next_sequence
       OR v_grant.next_sha256 IS DISTINCT FROM p_next_sha256
       OR v_grant.manifest_sha256 IS DISTINCT FROM p_manifest_sha256
       OR v_grant.authority_revision IS DISTINCT FROM p_authority_revision THEN
        RAISE EXCEPTION USING ERRCODE = '23505',
            MESSAGE = 'restored license scope input conflicts';
    END IF;
    IF v_grant.lease_until <= v_now THEN
        UPDATE platform.platformops_restored_license_grants AS grant_row
           SET lease_until = v_now + interval '60 seconds', updated_at = v_now
         WHERE grant_row.operation_id = p_operation_id
         RETURNING grant_row.* INTO v_grant;
    END IF;
    RETURN QUERY SELECT v_grant.operation_id, v_grant.environment_id,
        v_grant.restore_receipt_id, v_grant.restore_receipt_sha256,
        v_grant.license_binding_id, v_grant.license_id,
        v_grant.licensee_organization_id, v_grant.delivery_kind, v_grant.issuer,
        v_grant.previous_sequence, v_grant.previous_sha256,
        v_grant.next_sequence, v_grant.next_sha256, v_grant.manifest_sha256,
        v_grant.authority_revision, v_grant.lease_until, v_now;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_restored_license_scope_mint_v1(
    uuid,uuid,uuid,text,uuid,uuid,text,text,bigint,text,bigint,text,text,bigint
) FROM PUBLIC;
