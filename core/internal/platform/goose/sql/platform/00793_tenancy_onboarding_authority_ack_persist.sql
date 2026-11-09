-- +goose Up
DROP FUNCTION platform.tenancy_onboarding_authority_ack_v1(uuid, uuid, text, uuid, bigint, text, text, text, text, bigint, timestamptz, bigint);

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_authority_ack_v1(
    in_operation_id uuid, in_registry_binding_id uuid, in_scope_kind text, in_scope_id uuid, in_generation bigint,
    in_previous_head_sha256 text, in_state_sha256 text, in_proposal_sha256 text, in_ack_sha256 text,
    in_registry_sequence bigint, in_verified_at timestamptz, in_expected_revision bigint
) RETURNS TABLE (operation_id uuid, state text, revision bigint, updated_at timestamptz)
LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
DECLARE operation platform.tenancy_onboarding_operations%ROWTYPE; stamp timestamptz;
BEGIN
    IF in_ack_sha256 !~ '^[0-9a-f]{64}$' OR in_state_sha256 !~ '^[0-9a-f]{64}$' OR in_proposal_sha256 !~ '^[0-9a-f]{64}$'
        OR in_registry_sequence < 1 OR in_verified_at IS NULL THEN
        RAISE EXCEPTION 'authority acknowledgement is invalid' USING ERRCODE = '22023';
    END IF;
    SELECT * INTO operation FROM platform.tenancy_onboarding_operations WHERE operation_id = in_operation_id FOR UPDATE;
    IF NOT FOUND OR operation.registry_binding_id IS DISTINCT FROM in_registry_binding_id
        OR operation.authority_scope_kind IS DISTINCT FROM in_scope_kind OR operation.authority_scope_id IS DISTINCT FROM in_scope_id
        OR operation.authority_generation IS DISTINCT FROM in_generation THEN
        RAISE EXCEPTION 'authority acknowledgement does not match the operation' USING ERRCODE = '42501';
    END IF;
    IF EXISTS (SELECT 1 FROM platform.tenancy_onboarding_authority_acks WHERE operation_id = in_operation_id AND ack_sha256 = in_ack_sha256)
       AND operation.state = 'authority_acknowledged' THEN
        RETURN QUERY SELECT operation.operation_id, operation.state, operation.revision, operation.updated_at;
        RETURN;
    END IF;
    IF operation.state <> 'authority_pending' OR operation.revision <> in_expected_revision THEN
        RAISE EXCEPTION 'authority acknowledgement conflicts with the operation revision' USING ERRCODE = '40001';
    END IF;
    INSERT INTO platform.tenancy_onboarding_authority_acks (
        operation_id, registry_binding_id, scope_kind, scope_id, generation, previous_head_sha256,
        state_sha256, proposal_sha256, ack_sha256, registry_sequence, verified_at
    ) VALUES (
        in_operation_id, in_registry_binding_id, in_scope_kind, in_scope_id, in_generation,
        COALESCE(in_previous_head_sha256, ''), in_state_sha256, in_proposal_sha256, in_ack_sha256,
        in_registry_sequence, in_verified_at
    );
    stamp := pg_catalog.clock_timestamp();
    UPDATE platform.tenancy_onboarding_operations SET state = 'authority_acknowledged', revision = revision + 1, updated_at = stamp
    WHERE tenancy_onboarding_operations.operation_id = in_operation_id;
    RETURN QUERY SELECT in_operation_id, 'authority_acknowledged'::text, operation.revision + 1, stamp;
END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_onboarding_authority_ack_v1(uuid, uuid, text, uuid, bigint, text, text, text, text, bigint, timestamptz, bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_authority_ack_v1(uuid, uuid, text, uuid, bigint, text, text, text, text, bigint, timestamptz, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_onboarding_authority_ack_v1(uuid, uuid, text, uuid, bigint, text, text, text, text, bigint, timestamptz, bigint);
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_authority_ack_v1(
    in_operation_id uuid, in_registry_binding_id uuid, in_scope_kind text, in_scope_id uuid, in_generation bigint,
    in_previous_head_sha256 text, in_state_sha256 text, in_proposal_sha256 text, in_ack_sha256 text,
    in_registry_sequence bigint, in_verified_at timestamptz, in_expected_revision bigint
) RETURNS TABLE (operation_id uuid, state text, revision bigint, updated_at timestamptz)
LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$ SELECT in_operation_id, ''::text, 0::bigint, pg_catalog.clock_timestamp(); $fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.tenancy_onboarding_authority_ack_v1(uuid, uuid, text, uuid, bigint, text, text, text, text, bigint, timestamptz, bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_authority_ack_v1(uuid, uuid, text, uuid, bigint, text, text, text, text, bigint, timestamptz, bigint) TO harpy_api, harpy_worker;
