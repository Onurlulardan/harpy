-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_commit_restored_license_v1(
    p_operation_id uuid, p_restore_receipt_id uuid,
    p_restore_receipt_sha256 text, p_license_binding_id uuid,
    p_license_id uuid, p_licensee_organization_id uuid,
    p_environment_id uuid, p_previous_sequence bigint,
    p_previous_sha256 text, p_next_sequence bigint,
    p_next_sha256 text, p_expected_revision bigint,
    p_delivery_kind text, p_issuer text, p_key_id text, p_status text,
    p_issued_at timestamptz, p_not_before timestamptz,
    p_refresh_after timestamptz, p_grant_valid_until timestamptz,
    p_grace_until timestamptz, p_commercial_expires_at timestamptz,
    p_release_channel text, p_features jsonb, p_limits jsonb,
    p_canonical_payload bytea, p_signed_document bytea, p_projection jsonb
)
RETURNS TABLE (
    operation_id uuid, restore_receipt_id uuid, license_binding_id uuid,
    previous_sequence bigint, previous_sha256 text,
    accepted_sequence bigint, accepted_sha256 text,
    authority_revision bigint, replay boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_receipt platform.billing_license_restore_receipts%ROWTYPE;
    v_head platform.billing_license_heads%ROWTYPE;
    v_snapshot_id uuid;
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) <> 7
       OR p_restore_receipt_id IS NULL OR p_license_binding_id IS NULL
       OR p_restore_receipt_sha256 !~ '^[0-9a-f]{64}$'
       OR p_previous_sequence < 1 OR p_next_sequence <= p_previous_sequence
       OR p_previous_sha256 !~ '^[0-9a-f]{64}$'
       OR p_next_sha256 !~ '^[0-9a-f]{64}$'
       OR p_expected_revision < 1
       OR p_next_sha256 IS DISTINCT FROM pg_catalog.encode(
            pg_catalog.sha256(p_canonical_payload), 'hex')
       OR p_delivery_kind NOT IN ('connected', 'airgap', 'development')
       OR p_status NOT IN ('active', 'suspended', 'revoked', 'expired')
       OR pg_catalog.octet_length(p_issuer) NOT BETWEEN 1 AND 256
       OR p_key_id !~ '^[a-z0-9][a-z0-9._-]{0,63}$'
       OR p_release_channel !~ '^[a-z0-9][a-z0-9._-]{0,63}$'
       OR pg_catalog.jsonb_typeof(p_features) <> 'object'
       OR pg_catalog.jsonb_typeof(p_limits) <> 'object'
       OR pg_catalog.octet_length(p_signed_document) NOT BETWEEN 2 AND 65536 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'restored license commit is invalid';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'billing-restored-license|' || p_operation_id::text, 0
    ));
    SELECT receipt.* INTO v_receipt
    FROM platform.billing_license_restore_receipts AS receipt
    WHERE receipt.operation_id = p_operation_id;
    IF FOUND THEN
        IF v_receipt.restore_receipt_id IS DISTINCT FROM p_restore_receipt_id
           OR v_receipt.restore_receipt_sha256 IS DISTINCT FROM
                p_restore_receipt_sha256
           OR v_receipt.license_binding_id IS DISTINCT FROM p_license_binding_id
           OR v_receipt.license_id IS DISTINCT FROM p_license_id
           OR v_receipt.licensee_organization_id IS DISTINCT FROM
                p_licensee_organization_id
           OR v_receipt.environment_id IS DISTINCT FROM p_environment_id
           OR v_receipt.previous_sequence IS DISTINCT FROM p_previous_sequence
           OR v_receipt.previous_sha256 IS DISTINCT FROM p_previous_sha256
           OR v_receipt.accepted_sequence IS DISTINCT FROM p_next_sequence
           OR v_receipt.accepted_sha256 IS DISTINCT FROM p_next_sha256
           OR v_receipt.expected_revision IS DISTINCT FROM p_expected_revision THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'restored license operation was reused';
        END IF;
        RETURN QUERY SELECT p_operation_id, p_restore_receipt_id,
            p_license_binding_id, p_previous_sequence, p_previous_sha256,
            p_next_sequence, p_next_sha256,
            v_receipt.authority_revision, true;
        RETURN;
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        p_license_id::text || '|' || p_environment_id::text, 0
    ));
    SELECT head.* INTO v_head
    FROM platform.billing_license_heads AS head
    WHERE head.license_id = p_license_id
      AND head.environment_id = p_environment_id
    FOR UPDATE;
    IF NOT FOUND OR v_head.sequence IS DISTINCT FROM p_previous_sequence
       OR v_head.payload_sha256 IS DISTINCT FROM p_previous_sha256
       OR v_head.revision IS DISTINCT FROM p_expected_revision THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'restored license head compare-and-swap failed';
    END IF;
    INSERT INTO platform.billing_license_snapshots (
        license_id, licensee_organization_id, environment_id, delivery_kind,
        issuer, key_id, sequence, payload_sha256, status,
        issued_at, not_before, refresh_after, grant_valid_until,
        grace_until, commercial_expires_at, release_channel,
        features, limits, canonical_payload, signed_document, accepted_at
    ) VALUES (
        p_license_id, p_licensee_organization_id, p_environment_id,
        p_delivery_kind, p_issuer, p_key_id, p_next_sequence, p_next_sha256,
        p_status, p_issued_at, p_not_before, p_refresh_after,
        p_grant_valid_until, p_grace_until, p_commercial_expires_at,
        p_release_channel, p_features, p_limits,
        p_canonical_payload, p_signed_document, v_now
    ) RETURNING snapshot_id INTO v_snapshot_id;
    PERFORM platform.billing_license_projection_insert_v1(
        v_snapshot_id, p_projection, v_now
    );
    UPDATE platform.billing_license_heads
    SET snapshot_id = v_snapshot_id, sequence = p_next_sequence,
        payload_sha256 = p_next_sha256, revision = revision + 1,
        accepted_at = v_now
    WHERE license_id = p_license_id AND environment_id = p_environment_id;
    INSERT INTO platform.billing_license_restore_receipts (
        operation_id, restore_receipt_id, restore_receipt_sha256,
        license_binding_id, license_id, licensee_organization_id,
        environment_id, previous_sequence, previous_sha256,
        snapshot_id, accepted_sequence, accepted_sha256,
        expected_revision, authority_revision, created_at
    ) VALUES (
        p_operation_id, p_restore_receipt_id, p_restore_receipt_sha256,
        p_license_binding_id, p_license_id, p_licensee_organization_id,
        p_environment_id, p_previous_sequence, p_previous_sha256,
        v_snapshot_id, p_next_sequence, p_next_sha256,
        p_expected_revision, p_expected_revision + 1, v_now
    );
    RETURN QUERY SELECT p_operation_id, p_restore_receipt_id,
        p_license_binding_id, p_previous_sequence, p_previous_sha256,
        p_next_sequence, p_next_sha256, p_expected_revision + 1, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_commit_restored_license_v1(
    uuid, uuid, text, uuid, uuid, uuid, uuid, bigint, text, bigint, text,
    bigint, text, text, text, text, timestamptz, timestamptz, timestamptz,
    timestamptz, timestamptz, timestamptz, text, jsonb, jsonb,
    bytea, bytea, jsonb
) FROM PUBLIC;

