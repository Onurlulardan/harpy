-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_get_operation_v1(p_operation_id uuid)
RETURNS TABLE (
    operation_id uuid, purpose text, binding_id uuid, scope_kind text, scope_id uuid,
    generation bigint, previous_head_sha256 text, state_sha256 text, state_bytes bytea,
    proposal_sha256 text, proposal_bytes bytea, compensates_ack_sha256 text,
    status text, ack_sha256 text, ack_bytes bytea, registry_sequence bigint,
    completed_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY SELECT operation.operation_id, operation.purpose,
        operation.binding_id, operation.scope_kind, operation.scope_id,
        operation.generation, operation.previous_head_sha256, operation.state_sha256,
        operation.state_bytes, operation.proposal_sha256, operation.proposal_bytes,
        operation.compensates_ack_sha256, operation.status, operation.ack_sha256,
        operation.ack_bytes, operation.registry_sequence, operation.completed_at
    FROM platform.platformauthority_operations AS operation
    WHERE operation.operation_id = p_operation_id;
END;
$function$;

CREATE FUNCTION platform.platformauthority_get_head_v1(
    p_binding_id uuid, p_scope_kind text, p_scope_id uuid
)
RETURNS TABLE (
    binding_id uuid, scope_kind text, scope_id uuid, generation bigint,
    state_sha256 text, ack_sha256 text, registry_sequence bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY SELECT head.binding_id, head.scope_kind, head.scope_id,
        head.generation, head.state_sha256, head.ack_sha256, head.registry_sequence
    FROM platform.platformauthority_heads AS head
    WHERE head.binding_id = p_binding_id AND head.scope_kind = p_scope_kind
      AND head.scope_id = p_scope_id;
END;
$function$;
-- +goose StatementEnd
