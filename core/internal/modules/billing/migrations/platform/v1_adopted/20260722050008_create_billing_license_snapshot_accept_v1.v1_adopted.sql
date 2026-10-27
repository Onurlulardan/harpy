-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_license_snapshot_accept_v1(
    p_operation_id uuid, p_lease_token uuid,
    p_license_id uuid, p_licensee_organization_id uuid,
    p_environment_id uuid, p_delivery_kind text,
    p_issuer text, p_key_id text, p_payload_sha256 text,
    p_status text, p_sequence bigint,
    p_issued_at timestamptz, p_not_before timestamptz,
    p_refresh_after timestamptz, p_grant_valid_until timestamptz,
    p_grace_until timestamptz, p_commercial_expires_at timestamptz,
    p_release_channel text, p_features jsonb, p_limits jsonb,
    p_canonical_payload bytea, p_signed_document bytea,
    p_projection jsonb
)
RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_operation platform.billing_license_refresh_operations%ROWTYPE;
    v_head platform.billing_license_heads%ROWTYPE;
    v_head_found boolean;
    v_snapshot_id uuid;
    v_outcome text;
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF p_sequence < 1 OR p_payload_sha256 !~ '^[0-9a-f]{64}$'
       OR p_payload_sha256 IS DISTINCT FROM pg_catalog.encode(
            pg_catalog.sha256(p_canonical_payload), 'hex')
       OR pg_catalog.octet_length(p_signed_document) NOT BETWEEN 2 AND 65536
       OR p_delivery_kind NOT IN ('connected', 'airgap', 'development')
       OR p_status NOT IN ('active', 'suspended', 'revoked', 'expired')
       OR pg_catalog.octet_length(p_issuer) NOT BETWEEN 1 AND 256
       OR p_issuer <> pg_catalog.btrim(p_issuer)
       OR p_issuer ~ '[[:cntrl:]]'
       OR p_key_id !~ '^[a-z0-9][a-z0-9._-]{0,63}$'
       OR p_release_channel !~ '^[a-z0-9][a-z0-9._-]{0,63}$'
       OR pg_catalog.jsonb_typeof(p_features) <> 'object'
       OR pg_catalog.jsonb_typeof(p_limits) <> 'object'
       OR NOT pg_catalog.isfinite(p_issued_at)
       OR NOT pg_catalog.isfinite(p_not_before)
       OR NOT pg_catalog.isfinite(p_grant_valid_until)
       OR NOT pg_catalog.isfinite(p_grace_until) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing license snapshot acceptance is invalid';
    END IF;
    SELECT operation.* INTO v_operation
    FROM platform.billing_license_refresh_operations AS operation
    WHERE operation.operation_id = p_operation_id
    FOR UPDATE;
    IF NOT FOUND OR v_operation.state <> 'running'
       OR v_operation.lease_token IS DISTINCT FROM p_lease_token
       OR v_operation.lease_until <= v_now
       OR v_operation.license_id IS DISTINCT FROM p_license_id
       OR v_operation.licensee_organization_id IS DISTINCT FROM
            p_licensee_organization_id
       OR v_operation.environment_id IS DISTINCT FROM p_environment_id
       OR v_operation.delivery_kind IS DISTINCT FROM p_delivery_kind THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'billing license snapshot acceptance lease is stale';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        p_license_id::text || '|' || p_environment_id::text, 0
    ));
    SELECT head.* INTO v_head
    FROM platform.billing_license_heads AS head
    WHERE head.license_id = p_license_id
      AND head.environment_id = p_environment_id
    FOR UPDATE;
    v_head_found := FOUND;
    v_outcome := CASE
        WHEN NOT v_head_found THEN 'accepted'
        WHEN p_sequence < v_head.sequence THEN 'sequence_rollback'
        WHEN p_sequence = v_head.sequence
             AND p_payload_sha256 <> v_head.payload_sha256
            THEN 'sequence_conflict'
        WHEN p_sequence = v_head.sequence THEN 'noop'
        ELSE 'accepted'
    END;
    IF v_outcome <> 'accepted' THEN
        INSERT INTO platform.billing_license_verification_receipts (
            operation_id, attempt, license_id, environment_id,
            sequence, payload_sha256, outcome, verified_at
        ) VALUES (
            p_operation_id, v_operation.current_attempt, p_license_id,
            p_environment_id, p_sequence, p_payload_sha256, v_outcome, v_now
        );
        UPDATE platform.billing_license_refresh_attempts
        SET state = v_outcome, finished_at = v_now,
            failure_kind = CASE WHEN v_outcome = 'noop'
                THEN NULL ELSE v_outcome END
        WHERE operation_id = p_operation_id
          AND attempt = v_operation.current_attempt
          AND lease_token = p_lease_token AND state = 'running';
        UPDATE platform.billing_license_refresh_operations
        SET state = CASE WHEN v_outcome = 'noop'
                THEN 'succeeded' ELSE 'fenced' END,
            lease_token = NULL, lease_until = NULL,
            updated_at = v_now, completed_at = v_now
        WHERE operation_id = p_operation_id;
        RETURN v_outcome;
    END IF;
    INSERT INTO platform.billing_license_snapshots (
        license_id, licensee_organization_id, environment_id, delivery_kind,
        issuer, key_id, sequence, payload_sha256, status,
        issued_at, not_before, refresh_after, grant_valid_until,
        grace_until, commercial_expires_at, release_channel,
        features, limits, canonical_payload, signed_document, accepted_at
    ) VALUES (
        p_license_id, p_licensee_organization_id, p_environment_id,
        p_delivery_kind, p_issuer, p_key_id, p_sequence, p_payload_sha256,
        p_status, p_issued_at, p_not_before, p_refresh_after,
        p_grant_valid_until, p_grace_until, p_commercial_expires_at,
        p_release_channel, p_features, p_limits,
        p_canonical_payload, p_signed_document, v_now
    ) RETURNING snapshot_id INTO v_snapshot_id;
    PERFORM platform.billing_license_projection_insert_v1(
        v_snapshot_id, p_projection, v_now
    );
    IF v_head_found THEN
        UPDATE platform.billing_license_heads
        SET snapshot_id = v_snapshot_id, sequence = p_sequence,
            payload_sha256 = p_payload_sha256, revision = revision + 1,
            accepted_at = v_now
        WHERE license_id = p_license_id AND environment_id = p_environment_id;
    ELSE
        INSERT INTO platform.billing_license_heads (
            license_id, environment_id, snapshot_id, sequence,
            payload_sha256, revision, accepted_at
        ) VALUES (
            p_license_id, p_environment_id, v_snapshot_id, p_sequence,
            p_payload_sha256, 1, v_now
        );
    END IF;
    INSERT INTO platform.billing_license_verification_receipts (
        operation_id, attempt, license_id, environment_id,
        sequence, payload_sha256, outcome, verified_at
    ) VALUES (
        p_operation_id, v_operation.current_attempt, p_license_id,
        p_environment_id, p_sequence, p_payload_sha256, 'accepted', v_now
    );
    UPDATE platform.billing_license_refresh_attempts
    SET state = 'succeeded', finished_at = v_now
    WHERE operation_id = p_operation_id
      AND attempt = v_operation.current_attempt
      AND lease_token = p_lease_token AND state = 'running';
    UPDATE platform.billing_license_refresh_operations
    SET state = 'succeeded', lease_token = NULL, lease_until = NULL,
        updated_at = v_now, completed_at = v_now
    WHERE operation_id = p_operation_id;
    RETURN 'accepted';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_license_snapshot_accept_v1(
    uuid, uuid, uuid, uuid, uuid, text, text, text, text, text, bigint,
    timestamptz, timestamptz, timestamptz, timestamptz, timestamptz,
    timestamptz, text, jsonb, jsonb, bytea, bytea, jsonb
) FROM PUBLIC;

