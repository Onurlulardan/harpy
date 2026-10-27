-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_prepare_v1(
    p_operation_id uuid, p_binding_id uuid, p_purpose text, p_scope_kind text,
    p_scope_id uuid, p_generation bigint, p_previous_head_sha256 text,
    p_state_sha256 text, p_state_bytes bytea, p_proposal_sha256 text,
    p_proposal_bytes bytea, p_compensates_ack_sha256 text,
    p_issued_at timestamptz, p_expires_at timestamptz, p_now timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_generation bigint;
    v_ack_sha256 text;
    v_exact boolean;
BEGIN
    PERFORM 1 FROM platform.platformauthority_bindings WHERE binding_id = p_binding_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '23503', MESSAGE = 'protected authority binding is missing';
    END IF;
    SELECT generation, ack_sha256 INTO v_generation, v_ack_sha256
    FROM platform.platformauthority_heads
    WHERE binding_id = p_binding_id AND scope_kind = p_scope_kind AND scope_id = p_scope_id
    FOR UPDATE;
    IF (v_generation IS NULL AND (p_generation <> 1 OR p_previous_head_sha256 IS NOT NULL)) OR
       (v_generation IS NOT NULL AND (p_generation <> v_generation + 1 OR
            p_previous_head_sha256 IS DISTINCT FROM v_ack_sha256)) THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'protected authority head is stale';
    END IF;
    INSERT INTO platform.platformauthority_operations (
        operation_id, binding_id, purpose, scope_kind, scope_id, generation,
        previous_head_sha256, state_sha256, state_bytes, proposal_sha256,
        proposal_bytes, compensates_ack_sha256, status, issued_at, expires_at,
        created_at, updated_at
    ) VALUES (
        p_operation_id, p_binding_id, p_purpose, p_scope_kind, p_scope_id, p_generation,
        p_previous_head_sha256, p_state_sha256, p_state_bytes, p_proposal_sha256,
        p_proposal_bytes, p_compensates_ack_sha256, 'proposal_signed', p_issued_at,
        p_expires_at, p_now, p_now
    ) ON CONFLICT (operation_id) DO NOTHING;
    IF FOUND THEN
        RETURN;
    END IF;
    SELECT ROW(binding_id, purpose, scope_kind, scope_id, generation,
        previous_head_sha256, state_sha256, state_bytes, proposal_sha256,
        proposal_bytes, compensates_ack_sha256, issued_at, expires_at)
        IS NOT DISTINCT FROM
        ROW(p_binding_id, p_purpose, p_scope_kind, p_scope_id, p_generation,
        p_previous_head_sha256, p_state_sha256, p_state_bytes, p_proposal_sha256,
        p_proposal_bytes, p_compensates_ack_sha256, p_issued_at, p_expires_at)
    INTO v_exact FROM platform.platformauthority_operations
    WHERE operation_id = p_operation_id;
    IF NOT COALESCE(v_exact, false) THEN
        RAISE EXCEPTION USING ERRCODE = '23505', MESSAGE = 'protected authority operation conflicts';
    END IF;
END;
$function$;
-- +goose StatementEnd

