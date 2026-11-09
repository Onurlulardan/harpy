-- +goose Up
DROP FUNCTION platform.tenancy_onboarding_accept_v2(uuid, uuid, text, text, text, text, text, text, uuid, name, text, uuid, uuid, uuid, uuid, uuid, bigint, text, timestamptz, bigint, text, bigint, text, uuid, text, uuid, bigint, text, text, text, timestamptz);

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_accept_v2(
    in_operation_id uuid, in_environment_id uuid, in_kind text, in_idempotency_key text,
    in_request_sha256 text, in_manifest_sha256 text, in_step_plan_sha256 text, in_deployment_mode text,
    in_tenant_id uuid, in_schema_name name, in_binding_purpose text, in_first_user_id uuid,
    in_root_organization_id uuid, in_protected_admin_binding_id uuid, in_instance_owner_binding_id uuid,
    in_offering_version_id uuid, in_offering_revision bigint, in_offering_sha256 text,
    in_offering_effective_at timestamptz, in_source_sequence bigint, in_source_sha256 text,
    in_recovery_generation bigint, in_trust_bundle_sha256 text, in_registry_binding_id uuid,
    in_scope_kind text, in_scope_id uuid, in_authority_generation bigint, in_previous_head_sha256 text,
    in_state_sha256 text, in_proposal_sha256 text, in_created_at timestamptz
) RETURNS TABLE (operation_id uuid, state text, revision bigint, updated_at timestamptz)
LANGUAGE plpgsql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$
DECLARE existing platform.tenancy_onboarding_operations%ROWTYPE; next_state text; next_revision bigint; stamp timestamptz;
BEGIN
    stamp := CASE WHEN pg_catalog.clock_timestamp() >= in_created_at THEN pg_catalog.clock_timestamp() ELSE in_created_at END;
    SELECT * INTO existing FROM platform.tenancy_onboarding_operations WHERE tenancy_onboarding_operations.operation_id = in_operation_id;
    IF NOT FOUND THEN
        INSERT INTO platform.tenancy_onboarding_operations (
            operation_id, tenant_id, first_owner_user_id, kind, step_plan_sha256, deployment_mode, state, revision,
            attempt_number, updated_at, environment_id, idempotency_key, request_sha256, manifest_sha256, schema_name,
            binding_purpose, root_organization_id, protected_admin_binding_id, instance_owner_binding_id,
            offering_version_id, offering_revision, offering_sha256, offering_effective_at, offering_source_sequence,
            offering_source_sha256, recovery_generation, recovery_trust_bundle_sha256, registry_binding_id,
            authority_scope_kind, authority_scope_id, authority_generation, authority_previous_head_sha256,
            authority_state_sha256, authority_proposal_sha256, created_at
        ) VALUES (
            in_operation_id, in_tenant_id, in_first_user_id, in_kind, in_step_plan_sha256, in_deployment_mode,
            'accepted', 1, 0, stamp, in_environment_id, in_idempotency_key, in_request_sha256, in_manifest_sha256,
            in_schema_name, in_binding_purpose, in_root_organization_id, in_protected_admin_binding_id,
            in_instance_owner_binding_id, in_offering_version_id, in_offering_revision, in_offering_sha256,
            in_offering_effective_at, in_source_sequence, in_source_sha256, in_recovery_generation,
            in_trust_bundle_sha256, in_registry_binding_id, in_scope_kind, in_scope_id, in_authority_generation,
            in_previous_head_sha256, in_state_sha256, in_proposal_sha256, in_created_at
        );
        RETURN QUERY SELECT in_operation_id, 'accepted'::text, 1::bigint, stamp;
        RETURN;
    END IF;
    IF existing.tenant_id IS DISTINCT FROM in_tenant_id OR existing.first_owner_user_id IS DISTINCT FROM in_first_user_id
        OR existing.kind IS DISTINCT FROM in_kind OR existing.step_plan_sha256 IS DISTINCT FROM in_step_plan_sha256
        OR existing.deployment_mode IS DISTINCT FROM in_deployment_mode
        OR (existing.environment_id IS NOT NULL AND existing.environment_id IS DISTINCT FROM in_environment_id)
        OR (existing.manifest_sha256 IS NOT NULL AND existing.manifest_sha256 IS DISTINCT FROM in_manifest_sha256)
        OR (existing.schema_name IS NOT NULL AND existing.schema_name IS DISTINCT FROM in_schema_name::text)
        OR (existing.binding_purpose IS NOT NULL AND existing.binding_purpose IS DISTINCT FROM in_binding_purpose)
        OR (existing.idempotency_key IS NOT NULL AND existing.idempotency_key IS DISTINCT FROM in_idempotency_key)
        OR (existing.request_sha256 IS NOT NULL AND existing.request_sha256 IS DISTINCT FROM in_request_sha256)
        OR (existing.root_organization_id IS NOT NULL AND existing.root_organization_id IS DISTINCT FROM in_root_organization_id)
        OR (existing.protected_admin_binding_id IS NOT NULL AND existing.protected_admin_binding_id IS DISTINCT FROM in_protected_admin_binding_id)
        OR (existing.instance_owner_binding_id IS NOT NULL AND existing.instance_owner_binding_id IS DISTINCT FROM in_instance_owner_binding_id)
        OR (existing.offering_version_id IS NOT NULL AND existing.offering_version_id IS DISTINCT FROM in_offering_version_id)
        OR (existing.authority_generation IS NOT NULL AND existing.authority_generation IS DISTINCT FROM in_authority_generation) THEN
        RAISE EXCEPTION 'onboarding operation identity mismatch' USING ERRCODE = '23514';
    END IF;
    next_state := CASE existing.state
        WHEN 'accepted' THEN CASE WHEN in_deployment_mode = 'saas' THEN 'authority_pending' ELSE 'migration_ready' END
        WHEN 'migration_ready' THEN 'succeeded' WHEN 'authority_pending' THEN 'authority_acknowledged'
        WHEN 'authority_acknowledged' THEN 'succeeded' ELSE existing.state END;
    next_revision := CASE WHEN existing.state = 'succeeded' THEN existing.revision ELSE existing.revision + 1 END;
    UPDATE platform.tenancy_onboarding_operations SET
        state = next_state, revision = next_revision, updated_at = stamp,
        environment_id = COALESCE(environment_id, in_environment_id),
        idempotency_key = COALESCE(idempotency_key, in_idempotency_key),
        request_sha256 = COALESCE(request_sha256, in_request_sha256),
        manifest_sha256 = COALESCE(manifest_sha256, in_manifest_sha256),
        schema_name = COALESCE(schema_name, in_schema_name::text),
        binding_purpose = COALESCE(binding_purpose, in_binding_purpose),
        root_organization_id = COALESCE(root_organization_id, in_root_organization_id),
        protected_admin_binding_id = COALESCE(protected_admin_binding_id, in_protected_admin_binding_id),
        instance_owner_binding_id = COALESCE(instance_owner_binding_id, in_instance_owner_binding_id),
        offering_version_id = COALESCE(offering_version_id, in_offering_version_id),
        offering_revision = COALESCE(offering_revision, in_offering_revision),
        offering_sha256 = COALESCE(offering_sha256, in_offering_sha256),
        offering_effective_at = COALESCE(offering_effective_at, in_offering_effective_at),
        offering_source_sequence = COALESCE(offering_source_sequence, in_source_sequence),
        offering_source_sha256 = COALESCE(offering_source_sha256, in_source_sha256),
        recovery_generation = COALESCE(recovery_generation, in_recovery_generation),
        recovery_trust_bundle_sha256 = COALESCE(recovery_trust_bundle_sha256, in_trust_bundle_sha256),
        registry_binding_id = COALESCE(registry_binding_id, in_registry_binding_id),
        authority_scope_kind = COALESCE(authority_scope_kind, in_scope_kind),
        authority_scope_id = COALESCE(authority_scope_id, in_scope_id),
        authority_generation = COALESCE(authority_generation, in_authority_generation),
        authority_previous_head_sha256 = COALESCE(authority_previous_head_sha256, in_previous_head_sha256),
        authority_state_sha256 = COALESCE(authority_state_sha256, in_state_sha256),
        authority_proposal_sha256 = COALESCE(authority_proposal_sha256, in_proposal_sha256),
        created_at = COALESCE(created_at, in_created_at)
    WHERE tenancy_onboarding_operations.operation_id = in_operation_id;
    RETURN QUERY SELECT in_operation_id, next_state, next_revision, stamp;
END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_onboarding_accept_v2(uuid, uuid, text, text, text, text, text, text, uuid, name, text, uuid, uuid, uuid, uuid, uuid, bigint, text, timestamptz, bigint, text, bigint, text, uuid, text, uuid, bigint, text, text, text, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_accept_v2(uuid, uuid, text, text, text, text, text, text, uuid, name, text, uuid, uuid, uuid, uuid, uuid, bigint, text, timestamptz, bigint, text, bigint, text, uuid, text, uuid, bigint, text, text, text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_onboarding_accept_v2(uuid, uuid, text, text, text, text, text, text, uuid, name, text, uuid, uuid, uuid, uuid, uuid, bigint, text, timestamptz, bigint, text, bigint, text, uuid, text, uuid, bigint, text, text, text, timestamptz);
-- Restores the applied 00090 state machine. Identity columns stay until 00785 is rolled back.
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_accept_v2(
    in_operation_id uuid, in_environment_id uuid, in_kind text, in_idempotency_key text, in_request_sha256 text,
    in_manifest_sha256 text, in_step_plan_sha256 text, in_deployment_mode text, in_tenant_id uuid, in_schema_name name,
    in_binding_purpose text, in_first_user_id uuid, in_root_organization_id uuid, in_protected_admin_binding_id uuid,
    in_instance_owner_binding_id uuid, in_offering_version_id uuid, in_offering_revision bigint, in_offering_sha256 text,
    in_offering_effective_at timestamptz, in_source_sequence bigint, in_source_sha256 text, in_recovery_generation bigint,
    in_trust_bundle_sha256 text, in_registry_binding_id uuid, in_scope_kind text, in_scope_id uuid,
    in_authority_generation bigint, in_previous_head_sha256 text, in_state_sha256 text, in_proposal_sha256 text, in_created_at timestamptz
) RETURNS TABLE (operation_id uuid, state text, revision bigint, updated_at timestamptz)
LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$
    WITH clock AS (SELECT CASE WHEN pg_catalog.clock_timestamp() >= in_created_at THEN pg_catalog.clock_timestamp() ELSE in_created_at END AS updated_at),
    current_row AS (SELECT operation.state, operation.revision FROM platform.tenancy_onboarding_operations AS operation WHERE operation.operation_id = in_operation_id),
    nxt AS (SELECT CASE WHEN current_row.state IS NULL THEN 'accepted' WHEN current_row.state = 'accepted' AND in_deployment_mode = 'saas' THEN 'authority_pending' WHEN current_row.state = 'accepted' THEN 'migration_ready' WHEN current_row.state = 'migration_ready' THEN 'succeeded' WHEN current_row.state = 'authority_pending' THEN 'authority_acknowledged' WHEN current_row.state = 'authority_acknowledged' THEN 'succeeded' ELSE current_row.state END AS state, CASE WHEN current_row.state IS NULL THEN 1 WHEN current_row.state = 'succeeded' THEN current_row.revision ELSE current_row.revision + 1 END AS revision FROM clock LEFT JOIN current_row ON true),
    stored AS (INSERT INTO platform.tenancy_onboarding_operations (operation_id, tenant_id, first_owner_user_id, kind, step_plan_sha256, deployment_mode, state, revision, attempt_number, updated_at) SELECT in_operation_id, in_tenant_id, in_first_user_id, in_kind, in_step_plan_sha256, in_deployment_mode, nxt.state, nxt.revision, 0, clock.updated_at FROM nxt CROSS JOIN clock ON CONFLICT (operation_id) DO UPDATE SET state = EXCLUDED.state, revision = EXCLUDED.revision, updated_at = EXCLUDED.updated_at RETURNING operation_id, state, revision, updated_at)
    SELECT stored.operation_id, stored.state, stored.revision, stored.updated_at FROM stored;
$fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.tenancy_onboarding_accept_v2(uuid, uuid, text, text, text, text, text, text, uuid, name, text, uuid, uuid, uuid, uuid, uuid, bigint, text, timestamptz, bigint, text, bigint, text, uuid, text, uuid, bigint, text, text, text, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_accept_v2(uuid, uuid, text, text, text, text, text, text, uuid, name, text, uuid, uuid, uuid, uuid, uuid, bigint, text, timestamptz, bigint, text, bigint, text, uuid, text, uuid, bigint, text, text, text, timestamptz) TO harpy_api, harpy_worker;
