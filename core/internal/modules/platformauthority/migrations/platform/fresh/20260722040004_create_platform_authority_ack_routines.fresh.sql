-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_record_ack_v1(
    p_operation_id uuid, p_binding_id uuid, p_scope_kind text, p_scope_id uuid,
    p_generation bigint, p_previous_head_sha256 text, p_state_sha256 text,
    p_proposal_sha256 text, p_ack_sha256 text, p_ack_bytes bytea,
    p_registry_sequence bigint, p_verified_at timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_operation platform.platformauthority_operations%ROWTYPE;
BEGIN
    SELECT * INTO v_operation FROM platform.platformauthority_operations
    WHERE operation_id = p_operation_id FOR UPDATE;
    IF NOT FOUND OR ROW(v_operation.binding_id, v_operation.scope_kind,
        v_operation.scope_id, v_operation.generation, v_operation.previous_head_sha256,
        v_operation.state_sha256, v_operation.proposal_sha256)
        IS DISTINCT FROM ROW(p_binding_id, p_scope_kind, p_scope_id, p_generation,
        p_previous_head_sha256, p_state_sha256, p_proposal_sha256) THEN
        RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'protected authority ack mismatches proposal';
    END IF;
    IF v_operation.status IN ('ack_verified', 'completed') THEN
        IF ROW(v_operation.ack_sha256, v_operation.ack_bytes, v_operation.registry_sequence)
            IS DISTINCT FROM ROW(p_ack_sha256, p_ack_bytes, p_registry_sequence) THEN
            RAISE EXCEPTION USING ERRCODE = '23505', MESSAGE = 'protected authority ack replay conflicts';
        END IF;
        RETURN;
    END IF;
    UPDATE platform.platformauthority_operations
    SET status = 'ack_verified', ack_sha256 = p_ack_sha256, ack_bytes = p_ack_bytes,
        registry_sequence = p_registry_sequence, ack_verified_at = p_verified_at,
        revision = revision + 1, updated_at = p_verified_at
    WHERE operation_id = p_operation_id;
END;
$function$;

CREATE FUNCTION platform.platformauthority_commit_v1(
    p_operation_id uuid, p_completed_at timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_operation platform.platformauthority_operations%ROWTYPE;
    v_count integer;
    v_signer_status text;
    v_signer_kid text;
    v_proposal_signer_kid text;
BEGIN
    SELECT * INTO v_operation FROM platform.platformauthority_operations
    WHERE operation_id = p_operation_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002', MESSAGE = 'protected authority operation is missing';
    END IF;
    IF v_operation.status = 'completed' THEN
        RETURN;
    END IF;
    SELECT signer_status, signer_kid INTO v_signer_status, v_signer_kid
    FROM platform.platformauthority_bindings
    WHERE binding_id = v_operation.binding_id FOR UPDATE;
    v_proposal_signer_kid := convert_from(v_operation.proposal_bytes, 'UTF8')::jsonb->>'signer_kid';
    IF NOT FOUND OR v_signer_status <> 'active' OR
       v_signer_kid IS DISTINCT FROM v_proposal_signer_kid THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'protected authority signer is fenced';
    END IF;
    IF v_operation.status <> 'ack_verified' THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'protected authority ack is not verified';
    END IF;
    INSERT INTO platform.platformauthority_heads (
        binding_id, scope_kind, scope_id, generation, previous_head_sha256,
        state_sha256, proposal_sha256, ack_sha256, registry_sequence, updated_at
    ) VALUES (
        v_operation.binding_id, v_operation.scope_kind, v_operation.scope_id,
        v_operation.generation, v_operation.previous_head_sha256,
        v_operation.state_sha256, v_operation.proposal_sha256,
        v_operation.ack_sha256, v_operation.registry_sequence, p_completed_at
    ) ON CONFLICT (binding_id, scope_kind, scope_id) DO UPDATE SET
        generation = EXCLUDED.generation,
        previous_head_sha256 = EXCLUDED.previous_head_sha256,
        state_sha256 = EXCLUDED.state_sha256,
        proposal_sha256 = EXCLUDED.proposal_sha256,
        ack_sha256 = EXCLUDED.ack_sha256,
        registry_sequence = EXCLUDED.registry_sequence,
        updated_at = EXCLUDED.updated_at
    WHERE platform.platformauthority_heads.generation = EXCLUDED.generation - 1
      AND platform.platformauthority_heads.ack_sha256 = EXCLUDED.previous_head_sha256;
    GET DIAGNOSTICS v_count = ROW_COUNT;
    IF v_count <> 1 THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'protected authority head CAS is stale';
    END IF;
    UPDATE platform.platformauthority_operations
    SET status = 'completed', completed_at = p_completed_at,
        revision = revision + 1, updated_at = p_completed_at
    WHERE operation_id = p_operation_id;
    INSERT INTO platform.platformauthority_receipts (
        operation_id, binding_id, scope_kind, scope_id, generation, state_sha256,
        proposal_sha256, ack_sha256, registry_sequence, completed_at
    ) VALUES (
        v_operation.operation_id, v_operation.binding_id, v_operation.scope_kind,
        v_operation.scope_id, v_operation.generation, v_operation.state_sha256,
        v_operation.proposal_sha256, v_operation.ack_sha256,
        v_operation.registry_sequence, p_completed_at
    );
END;
$function$;
-- +goose StatementEnd
