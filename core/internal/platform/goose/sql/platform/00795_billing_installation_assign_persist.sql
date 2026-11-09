-- +goose Up
CREATE TABLE IF NOT EXISTS platform.billing_installation_base_assignments (
    assignment_id uuid PRIMARY KEY,
    operation_id uuid NOT NULL UNIQUE,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    authority_id uuid NOT NULL UNIQUE,
    deployment_mode text NOT NULL,
    binding_purpose text NOT NULL,
    offering_version_id uuid NOT NULL,
    offering_revision bigint NOT NULL,
    offering_sha256 text NOT NULL,
    offering_effective_at timestamptz NOT NULL,
    source_sequence bigint,
    source_sha256 text,
    manifest_sha256 text NOT NULL,
    authority_generation bigint NOT NULL,
    created_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS platform.billing_installation_usage_openings (
    operation_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    manifest_sha256 text NOT NULL,
    meter_count integer NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT billing_installation_usage_openings_meter_check CHECK (meter_count > 0)
);

REVOKE ALL ON TABLE platform.billing_installation_base_assignments FROM PUBLIC, harpy_api, harpy_worker;
REVOKE ALL ON TABLE platform.billing_installation_usage_openings FROM PUBLIC, harpy_api, harpy_worker;

DROP FUNCTION platform.billing_installation_assign_v3(uuid, uuid, uuid, uuid, uuid, text, uuid, bigint, text, timestamptz, bigint, text, text, text);
DROP FUNCTION platform.billing_installation_open_usage_v2(uuid, uuid, text);

-- +goose StatementBegin
CREATE FUNCTION platform.billing_installation_assign_v3(
    in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_authority_id uuid, in_assignment_id uuid,
    in_deployment_mode text, in_offering_version_id uuid, in_offering_revision bigint, in_offering_sha256 text,
    in_offering_effective_at timestamptz, in_source_sequence bigint, in_source_sha256 text, in_manifest_sha256 text, in_binding_purpose text
) RETURNS boolean LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
DECLARE operation platform.tenancy_onboarding_operations%ROWTYPE; stored platform.billing_installation_base_assignments%ROWTYPE; generation bigint;
BEGIN
    SELECT * INTO operation FROM platform.tenancy_onboarding_operations WHERE operation_id = in_operation_id;
    IF NOT FOUND OR operation.tenant_id <> in_tenant_id OR operation.environment_id IS DISTINCT FROM in_environment_id
        OR operation.deployment_mode IS DISTINCT FROM in_deployment_mode OR operation.binding_purpose IS DISTINCT FROM in_binding_purpose
        OR operation.manifest_sha256 IS DISTINCT FROM in_manifest_sha256 OR operation.offering_version_id IS DISTINCT FROM in_offering_version_id
        OR operation.schema_name IS NULL OR NOT platform.tenancy_activation_assert_v2(in_operation_id, in_tenant_id, operation.schema_name::name) THEN
        RAISE EXCEPTION 'billing installation assignment does not match the activation' USING ERRCODE = '55000';
    END IF;
    generation := operation.authority_generation;
    IF EXISTS (SELECT 1 FROM platform.billing_tenant_authorities WHERE authority_id = in_authority_id AND tenant_id = in_tenant_id
        AND (billing_generation IS DISTINCT FROM generation OR authority_group_generation IS DISTINCT FROM generation)) THEN
        RAISE EXCEPTION 'billing installation authority generation is stale' USING ERRCODE = 'P5721';
    END IF;
    SELECT * INTO stored FROM platform.billing_installation_base_assignments WHERE operation_id = in_operation_id;
    IF FOUND THEN
        IF stored.assignment_id = in_assignment_id AND stored.authority_id = in_authority_id AND stored.tenant_id = in_tenant_id
           AND stored.manifest_sha256 = in_manifest_sha256 AND stored.authority_generation = generation THEN
            RETURN true;
        END IF;
        RAISE EXCEPTION 'billing installation assignment does not match the stored assignment' USING ERRCODE = '23514';
    END IF;
    IF NOT platform.tenancy_activation_step_context_assert_v1(in_operation_id, 'billing.assign_tenant_base_offering', in_tenant_id, operation.schema_name::name) THEN
        RAISE EXCEPTION 'billing installation assignment capability is missing' USING ERRCODE = '42501';
    END IF;
    INSERT INTO platform.billing_tenant_authorities (
        authority_id, scope_kind, scope_id, authority_group_id, billing_domain, tenant_id, authority_group_generation, billing_generation
    ) VALUES (
        in_authority_id, 'tenant', in_tenant_id, in_operation_id, 'installation', in_tenant_id, generation, generation
    );
    INSERT INTO platform.billing_installation_base_assignments (
        assignment_id, operation_id, environment_id, tenant_id, authority_id, deployment_mode, binding_purpose,
        offering_version_id, offering_revision, offering_sha256, offering_effective_at, source_sequence, source_sha256,
        manifest_sha256, authority_generation, created_at
    ) VALUES (
        in_assignment_id, in_operation_id, in_environment_id, in_tenant_id, in_authority_id, in_deployment_mode, in_binding_purpose,
        in_offering_version_id, in_offering_revision, in_offering_sha256, in_offering_effective_at, in_source_sequence, in_source_sha256,
        in_manifest_sha256, generation, pg_catalog.clock_timestamp()
    );
    PERFORM platform.tenancy_activation_step_consume_v1(in_operation_id, 'billing.assign_tenant_base_offering', in_tenant_id, operation.schema_name::name);
    RETURN false;
END;
$fn$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION platform.billing_installation_open_usage_v2(in_operation_id uuid, in_tenant_id uuid, in_manifest_sha256 text)
RETURNS TABLE (meter_count integer, replay boolean)
LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
DECLARE operation platform.tenancy_onboarding_operations%ROWTYPE; opened integer; authority uuid; definition_id uuid; period_id uuid;
BEGIN
    SELECT * INTO operation FROM platform.tenancy_onboarding_operations WHERE operation_id = in_operation_id;
    IF NOT FOUND OR operation.tenant_id <> in_tenant_id OR operation.manifest_sha256 IS DISTINCT FROM in_manifest_sha256 OR operation.schema_name IS NULL THEN
        RAISE EXCEPTION 'billing installation usage does not match the activation' USING ERRCODE = '55000';
    END IF;
    SELECT billing_installation_usage_openings.meter_count INTO opened FROM platform.billing_installation_usage_openings
    WHERE operation_id = in_operation_id AND tenant_id = in_tenant_id AND manifest_sha256 = in_manifest_sha256;
    IF FOUND THEN RETURN QUERY SELECT opened, true; RETURN; END IF;
    IF NOT platform.tenancy_activation_step_context_assert_v1(in_operation_id, 'billing.open_tenant_usage', in_tenant_id, operation.schema_name::name) THEN
        RAISE EXCEPTION 'billing installation usage capability is missing' USING ERRCODE = '42501';
    END IF;
    SELECT authority_id INTO authority FROM platform.billing_installation_base_assignments WHERE operation_id = in_operation_id AND tenant_id = in_tenant_id;
    IF authority IS NULL THEN RAISE EXCEPTION 'billing installation usage has no assignment' USING ERRCODE = '23514'; END IF;
    definition_id := platform.tenancy_activation_digest_uuid_v1('billing.entitlement|iam.active_users');
    period_id := platform.tenancy_activation_digest_uuid_v1('billing.usage.period|' || in_operation_id::text);
    INSERT INTO platform.billing_entitlement_definitions (
        id, entitlement_key, value_type, meter_kind, reset_policy, enforcement, unlimited_allowed
    ) VALUES (
        definition_id, 'iam.active_users', 'integer', 'gauge', 'never', 'hard', 'true'
    ) ON CONFLICT (id) DO NOTHING;
    INSERT INTO platform.billing_meter_periods (
        id, authority_id, tenant_id, entitlement_definition_id, entitlement_key, period_start, period_end, state, meter_kind
    ) VALUES (
        period_id, authority, in_tenant_id, definition_id, 'iam.active_users', pg_catalog.clock_timestamp()::text, 'never', 'open', 'gauge'
    );
    INSERT INTO platform.billing_usage_totals (
        reserved_total, revision, updated_at, authority_id, tenant_id, meter_period_id, entitlement_key, meter_kind, committed_total, gauge_total
    ) VALUES (
        in_tenant_id::text || ':iam.active_users', 1, pg_catalog.clock_timestamp(), authority, in_tenant_id, period_id, 'iam.active_users', 'gauge', '0', '0'
    );
    INSERT INTO platform.billing_usage_entries (
        id, authority_id, tenant_id, decision_id, entitlement_key, meter_period_id, reservation_id, entry_kind,
        amount_delta, committed_after, reserved_after, gauge_after, source_operation_id, owner_fact_id, evidence_sha256, occurred_at, created_at
    ) VALUES (
        platform.tenancy_activation_digest_uuid_v1('billing.usage.entry|' || in_operation_id::text), authority, in_tenant_id,
        platform.tenancy_activation_digest_uuid_v1('billing.usage.decision|' || in_operation_id::text), 'iam.active_users', period_id,
        platform.tenancy_activation_digest_uuid_v1('billing.usage.reservation|' || in_operation_id::text), 'open',
        '0', '0', '0', '0', in_operation_id, in_operation_id, operation.manifest_sha256, pg_catalog.clock_timestamp(), pg_catalog.clock_timestamp()
    );
    INSERT INTO platform.billing_installation_usage_openings (operation_id, tenant_id, manifest_sha256, meter_count, created_at)
    VALUES (in_operation_id, in_tenant_id, in_manifest_sha256, 1, pg_catalog.clock_timestamp());
    PERFORM platform.tenancy_activation_step_consume_v1(in_operation_id, 'billing.open_tenant_usage', in_tenant_id, operation.schema_name::name);
    RETURN QUERY SELECT 1, false;
END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_installation_assign_v3(uuid, uuid, uuid, uuid, uuid, text, uuid, bigint, text, timestamptz, bigint, text, text, text) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.billing_installation_open_usage_v2(uuid, uuid, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_installation_assign_v3(uuid, uuid, uuid, uuid, uuid, text, uuid, bigint, text, timestamptz, bigint, text, text, text) TO harpy_api, harpy_worker;
GRANT EXECUTE ON FUNCTION platform.billing_installation_open_usage_v2(uuid, uuid, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_installation_open_usage_v2(uuid, uuid, text);
DROP FUNCTION platform.billing_installation_assign_v3(uuid, uuid, uuid, uuid, uuid, text, uuid, bigint, text, timestamptz, bigint, text, text, text);
DROP TABLE IF EXISTS platform.billing_installation_usage_openings;
DROP TABLE IF EXISTS platform.billing_installation_base_assignments;
-- +goose StatementBegin
CREATE FUNCTION platform.billing_installation_assign_v3(in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_authority_id uuid, in_assignment_id uuid, in_deployment_mode text, in_offering_version_id uuid, in_offering_revision bigint, in_offering_sha256 text, in_offering_effective_at timestamptz, in_source_sequence bigint, in_source_sha256 text, in_manifest_sha256 text, in_binding_purpose text)
RETURNS boolean LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT false; $fn$;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE FUNCTION platform.billing_installation_open_usage_v2(in_operation_id uuid, in_tenant_id uuid, in_manifest_sha256 text)
RETURNS TABLE (meter_count integer, replay boolean) LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT 0::integer, false; $fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.billing_installation_assign_v3(uuid, uuid, uuid, uuid, uuid, text, uuid, bigint, text, timestamptz, bigint, text, text, text) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.billing_installation_open_usage_v2(uuid, uuid, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_installation_assign_v3(uuid, uuid, uuid, uuid, uuid, text, uuid, bigint, text, timestamptz, bigint, text, text, text) TO harpy_api, harpy_worker;
GRANT EXECUTE ON FUNCTION platform.billing_installation_open_usage_v2(uuid, uuid, text) TO harpy_api, harpy_worker;
