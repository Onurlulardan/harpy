-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_get_binding_v2(p_binding_id uuid)
RETURNS TABLE (
    binding_id uuid, environment_id uuid, deployment_mode text, registry_kind text,
    license_binding_id uuid, finalization_receipt_sha256 text,
    recovery_trust_bundle_sha256 text,
    signer_generation bigint, signer_kid text, signer_public_key_sha256 text,
    signer_status text
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY SELECT binding.binding_id, binding.environment_id,
        binding.deployment_mode, binding.registry_kind, binding.license_binding_id,
        binding.finalization_receipt_sha256, binding.recovery_trust_bundle_sha256,
        binding.signer_generation,
        binding.signer_kid, binding.signer_public_key_sha256, binding.signer_status
    FROM platform.platformauthority_bindings AS binding
    WHERE binding.binding_id = p_binding_id;
END;
$function$;

CREATE FUNCTION platform.platformauthority_bind_v2(
    p_binding_id uuid, p_environment_id uuid, p_deployment_mode text,
    p_registry_kind text, p_license_binding_id uuid,
    p_finalization_receipt_sha256 text, p_recovery_trust_bundle_sha256 text,
    p_signer_generation bigint, p_signer_kid text, p_signer_public_key_sha256 text,
    p_proof_sha256 text, p_proof_bytes bytea, p_created_at timestamptz
)
RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    PERFORM platform.platformauthority_bind_v1(
        p_binding_id, p_environment_id, p_deployment_mode, p_registry_kind,
        p_license_binding_id, p_finalization_receipt_sha256,
        p_recovery_trust_bundle_sha256, p_signer_generation, p_signer_kid,
        p_signer_public_key_sha256, p_proof_sha256, p_proof_bytes, p_created_at
    );
    INSERT INTO platform.platformauthority_signer_events (
        binding_id, signer_generation, signer_kid, signer_public_key_sha256,
        event_kind, occurred_at
    ) VALUES (p_binding_id, p_signer_generation, p_signer_kid,
        p_signer_public_key_sha256, 'activated', p_created_at)
    ON CONFLICT (binding_id, signer_generation, event_kind) DO NOTHING;
END;
$function$;

CREATE FUNCTION platform.platformauthority_prepare_v2(
    p_operation_id uuid, p_binding_id uuid, p_purpose text, p_scope_kind text,
    p_scope_id uuid, p_generation bigint, p_previous_head_sha256 text,
    p_state_sha256 text, p_state_bytes bytea, p_proposal_sha256 text,
    p_proposal_bytes bytea, p_compensates_ack_sha256 text,
    p_issued_at timestamptz, p_expires_at timestamptz, p_signer_kid text,
    p_now timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_kid text;
    v_status text;
BEGIN
    SELECT signer_kid, signer_status INTO v_kid, v_status
    FROM platform.platformauthority_bindings WHERE binding_id = p_binding_id FOR UPDATE;
    IF NOT FOUND OR v_status <> 'active' OR v_kid <> p_signer_kid THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'protected authority signer is fenced';
    END IF;
    PERFORM platform.platformauthority_prepare_v1(
        p_operation_id, p_binding_id, p_purpose, p_scope_kind, p_scope_id,
        p_generation, p_previous_head_sha256, p_state_sha256, p_state_bytes,
        p_proposal_sha256, p_proposal_bytes, p_compensates_ack_sha256,
        p_issued_at, p_expires_at, p_now
    );
END;
$function$;

CREATE FUNCTION platform.platformauthority_rotate_signer_v1(
    p_operation_id uuid, p_binding_id uuid, p_expected_generation bigint,
    p_expected_kid text, p_new_generation bigint, p_new_kid text,
    p_new_public_key_sha256 text, p_new_recovery_trust_bundle_sha256 text,
    p_now timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_binding platform.platformauthority_bindings%ROWTYPE;
    v_operation platform.platformauthority_operations%ROWTYPE;
    v_head_ack text;
BEGIN
    SELECT * INTO v_binding FROM platform.platformauthority_bindings
    WHERE binding_id = p_binding_id FOR UPDATE;
    IF v_binding.signer_generation = p_new_generation AND v_binding.signer_kid = p_new_kid
       AND v_binding.signer_public_key_sha256 = p_new_public_key_sha256
       AND v_binding.recovery_trust_bundle_sha256 IS NOT DISTINCT FROM
           p_new_recovery_trust_bundle_sha256
       AND v_binding.signer_status = 'active' THEN
        RETURN;
    END IF;
    IF NOT FOUND OR v_binding.signer_generation <> p_expected_generation OR
       v_binding.signer_kid <> p_expected_kid OR p_new_generation <> p_expected_generation + 1 OR
       p_new_kid = p_expected_kid OR v_binding.signer_status <> 'active' THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'protected authority signer rotation is stale';
    END IF;
    IF (v_binding.deployment_mode = 'onpremise' AND
            (p_new_recovery_trust_bundle_sha256 IS NULL OR
             p_new_recovery_trust_bundle_sha256 = v_binding.recovery_trust_bundle_sha256)) OR
       (v_binding.deployment_mode = 'saas' AND p_new_recovery_trust_bundle_sha256 IS NOT NULL) THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'signer rotation trust bundle conflicts';
    END IF;
    SELECT * INTO v_operation FROM platform.platformauthority_operations
    WHERE operation_id = p_operation_id AND binding_id = p_binding_id FOR UPDATE;
    SELECT ack_sha256 INTO v_head_ack FROM platform.platformauthority_heads
    WHERE binding_id = p_binding_id AND scope_kind = v_operation.scope_kind
      AND scope_id = v_operation.scope_id FOR UPDATE;
    IF v_operation.operation_id IS NULL OR v_head_ack IS NULL OR
       v_operation.status <> 'completed' OR v_operation.purpose <> 'proposal_signer_rotation'
       OR v_operation.ack_sha256 IS DISTINCT FROM v_head_ack THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'signer rotation ack is not current';
    END IF;
    UPDATE platform.platformauthority_bindings
    SET signer_generation = p_new_generation, signer_kid = p_new_kid,
        signer_public_key_sha256 = p_new_public_key_sha256,
        recovery_trust_bundle_sha256 = p_new_recovery_trust_bundle_sha256,
        signer_status = 'active'
    WHERE binding_id = p_binding_id;
    INSERT INTO platform.platformauthority_signer_events (
        binding_id, signer_generation, signer_kid, signer_public_key_sha256,
        event_kind, evidence_sha256, occurred_at
    ) VALUES
        (p_binding_id, v_binding.signer_generation, v_binding.signer_kid,
            v_binding.signer_public_key_sha256, 'retired', v_operation.ack_sha256, p_now),
        (p_binding_id, p_new_generation, p_new_kid, p_new_public_key_sha256,
            'activated', v_operation.ack_sha256, p_now);
END;
$function$;

CREATE FUNCTION platform.platformauthority_fence_signer_v1(
    p_binding_id uuid, p_expected_generation bigint, p_expected_kid text,
    p_evidence_sha256 text, p_now timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_binding platform.platformauthority_bindings%ROWTYPE;
BEGIN
    SELECT * INTO v_binding FROM platform.platformauthority_bindings
    WHERE binding_id = p_binding_id FOR UPDATE;
    IF v_binding.signer_status = 'rekey_required' AND
       v_binding.signer_generation = p_expected_generation AND v_binding.signer_kid = p_expected_kid THEN
        RETURN;
    END IF;
    IF NOT FOUND OR v_binding.signer_status <> 'active' OR
       v_binding.signer_generation <> p_expected_generation OR v_binding.signer_kid <> p_expected_kid THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'protected authority signer fence is stale';
    END IF;
    UPDATE platform.platformauthority_bindings SET signer_status = 'rekey_required'
    WHERE binding_id = p_binding_id;
    INSERT INTO platform.platformauthority_signer_events (
        binding_id, signer_generation, signer_kid, signer_public_key_sha256,
        event_kind, evidence_sha256, occurred_at
    ) VALUES (p_binding_id, v_binding.signer_generation, v_binding.signer_kid,
        v_binding.signer_public_key_sha256, 'revoked', p_evidence_sha256, p_now);
END;
$function$;
-- +goose StatementEnd
