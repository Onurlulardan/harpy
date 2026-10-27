-- +goose Up
-- Migration stream version: 20260729100032.
ALTER TABLE platform.platformauthority_installation_head_pins
    DROP CONSTRAINT platformauthority_installation_head_pins_scope_kind_check;
ALTER TABLE platform.platformauthority_installation_head_pins
    ADD CONSTRAINT platformauthority_installation_head_pins_scope_kind_check
    CHECK (scope_kind IN ('environment', 'tenant'));

-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_installation_head_pin_v3(
    p_operation_id uuid, p_environment_id uuid, p_tenant_id uuid,
    p_schema_name name, p_registry_binding_id uuid, p_scope_kind text,
    p_scope_id uuid, p_generation bigint, p_previous_head_sha256 text,
    p_state_sha256 text, p_proposal_sha256 text, p_ack_sha256 text,
    p_manifest_sha256 text
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_authority_operation platform.platformauthority_operations%ROWTYPE;
    v_head platform.platformauthority_heads%ROWTYPE;
    v_existing platform.platformauthority_installation_head_pins%ROWTYPE;
    v_operation_kind text;
    v_purpose text;
    v_replayed boolean := false;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR p_operation_id IS NULL OR p_environment_id IS NULL
       OR p_tenant_id IS NULL OR p_schema_name IS NULL
       OR p_registry_binding_id IS NULL OR p_scope_kind IS NULL
       OR p_scope_id IS NULL
       OR p_scope_kind NOT IN ('environment', 'tenant')
       OR (p_scope_kind = 'environment'
           AND p_scope_id IS DISTINCT FROM p_environment_id)
       OR (p_scope_kind = 'tenant'
           AND p_scope_id IS DISTINCT FROM p_tenant_id)
       OR p_generation IS NULL OR p_generation < 1
       OR ((p_generation = 1) IS DISTINCT FROM
           (p_previous_head_sha256 IS NULL))
       OR (p_previous_head_sha256 IS NOT NULL
           AND p_previous_head_sha256 !~ '^[0-9a-f]{64}$')
       OR p_state_sha256 IS NULL
       OR p_state_sha256 !~ '^[0-9a-f]{64}$'
       OR p_proposal_sha256 IS NULL
       OR p_proposal_sha256 !~ '^[0-9a-f]{64}$'
       OR p_ack_sha256 IS NULL
       OR p_ack_sha256 !~ '^[0-9a-f]{64}$'
       OR p_manifest_sha256 IS NULL
       OR p_manifest_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'installation authority head pin denied';
    END IF;
    v_operation_kind := CASE p_scope_kind
        WHEN 'environment' THEN 'environment_initialize'
        ELSE 'saas_customer_tenant_provision'
    END;
    v_purpose := CASE p_scope_kind
        WHEN 'environment' THEN 'initial_environment_authority'
        ELSE 'initial_tenant_authority'
    END;
    IF NOT platform.tenancy_activation_step_context_assert_v1(
        p_operation_id,
        'platformauthority.pin_protected_authority_head',
        p_environment_id, p_tenant_id, p_schema_name,
        v_operation_kind
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'installation authority head capability required';
    END IF;
    SELECT * INTO v_authority_operation
    FROM platform.platformauthority_operations AS authority_operation
    WHERE authority_operation.ack_sha256 = p_ack_sha256
    FOR UPDATE;
    IF NOT FOUND
       OR v_authority_operation.binding_id
            IS DISTINCT FROM p_registry_binding_id
       OR v_authority_operation.purpose IS DISTINCT FROM v_purpose
       OR v_authority_operation.scope_kind IS DISTINCT FROM p_scope_kind
       OR v_authority_operation.scope_id IS DISTINCT FROM p_scope_id
       OR v_authority_operation.generation IS DISTINCT FROM p_generation
       OR v_authority_operation.previous_head_sha256
            IS DISTINCT FROM p_previous_head_sha256
       OR v_authority_operation.state_sha256 IS DISTINCT FROM p_state_sha256
       OR v_authority_operation.proposal_sha256
            IS DISTINCT FROM p_proposal_sha256
       OR v_authority_operation.status NOT IN ('ack_verified', 'completed') THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'verified protected authority operation conflicts';
    END IF;
    IF v_authority_operation.status = 'ack_verified' THEN
        PERFORM platform.platformauthority_commit_v1(
            v_authority_operation.operation_id,
            pg_catalog.statement_timestamp()
        );
    END IF;
    SELECT * INTO v_head
    FROM platform.platformauthority_heads AS head
    WHERE head.binding_id = p_registry_binding_id
      AND head.scope_kind = p_scope_kind AND head.scope_id = p_scope_id
    FOR SHARE;
    IF NOT FOUND OR ROW(v_head.generation, v_head.previous_head_sha256,
        v_head.state_sha256, v_head.proposal_sha256, v_head.ack_sha256)
       IS DISTINCT FROM ROW(p_generation, p_previous_head_sha256,
        p_state_sha256, p_proposal_sha256, p_ack_sha256) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'protected authority head conflicts';
    END IF;
    SELECT * INTO v_existing
    FROM platform.platformauthority_installation_head_pins AS pin
    WHERE pin.operation_id = p_operation_id
    FOR SHARE;
    IF FOUND THEN
        IF ROW(v_existing.registry_binding_id, v_existing.scope_kind,
            v_existing.scope_id, v_existing.generation,
            v_existing.previous_head_sha256, v_existing.state_sha256,
            v_existing.proposal_sha256, v_existing.ack_sha256,
            v_existing.manifest_sha256)
           IS DISTINCT FROM ROW(p_registry_binding_id, p_scope_kind,
            p_scope_id, p_generation, p_previous_head_sha256,
            p_state_sha256, p_proposal_sha256, p_ack_sha256,
            p_manifest_sha256) THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'installation authority head replay conflicts';
        END IF;
        v_replayed := true;
    ELSE
        INSERT INTO platform.platformauthority_installation_head_pins (
            operation_id, registry_binding_id, scope_kind, scope_id,
            generation, previous_head_sha256, state_sha256,
            proposal_sha256, ack_sha256, manifest_sha256, pinned_at
        ) VALUES (
            p_operation_id, p_registry_binding_id, p_scope_kind, p_scope_id,
            p_generation, p_previous_head_sha256, p_state_sha256,
            p_proposal_sha256, p_ack_sha256, p_manifest_sha256,
            pg_catalog.statement_timestamp()
        );
    END IF;
    IF NOT platform.tenancy_activation_step_consume_v1(
        p_operation_id,
        'platformauthority.pin_protected_authority_head'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'installation authority head capability consume failed';
    END IF;
    RETURN v_replayed;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platformauthority_installation_head_pin_v3(
        uuid,uuid,uuid,name,uuid,text,uuid,bigint,text,text,text,text,text
    ) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
    platform.platformauthority_installation_head_pin_v3(
        uuid,uuid,uuid,name,uuid,text,uuid,bigint,text,text,text,text,text
    ) TO harpy_runtime;
