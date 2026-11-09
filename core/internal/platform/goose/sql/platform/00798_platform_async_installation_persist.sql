-- +goose Up
CREATE TABLE IF NOT EXISTS platform.platform_async_installation_activation_receipts (
    operation_id uuid NOT NULL,
    receipt_kind text NOT NULL,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    event_id uuid,
    artifact_id uuid,
    manifest_sha256 text NOT NULL,
    created_at timestamptz NOT NULL,
    PRIMARY KEY (operation_id, receipt_kind)
);

REVOKE ALL ON TABLE platform.platform_async_installation_activation_receipts FROM PUBLIC, harpy_api, harpy_worker;

DROP FUNCTION platform.platform_async_installation_activation_event_v2(uuid, uuid, uuid, uuid, name, text);
DROP FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(uuid, uuid, uuid, uuid, text, name, text);

-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_installation_activation_event_v2(
    in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_actor_id uuid, in_schema_name name, in_manifest_sha256 text
) RETURNS boolean LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
DECLARE operation platform.tenancy_onboarding_operations%ROWTYPE; recorded_event uuid; payload_sha text; inserted boolean := false;
BEGIN
    SELECT * INTO operation FROM platform.tenancy_onboarding_operations WHERE operation_id = in_operation_id;
    IF NOT FOUND OR operation.tenant_id <> in_tenant_id OR operation.environment_id IS DISTINCT FROM in_environment_id
        OR operation.schema_name IS DISTINCT FROM in_schema_name::text OR operation.manifest_sha256 IS DISTINCT FROM in_manifest_sha256
        OR operation.first_owner_user_id IS DISTINCT FROM in_actor_id THEN
        RAISE EXCEPTION 'installation activation event does not match the operation' USING ERRCODE = '42501';
    END IF;
    IF EXISTS (SELECT 1 FROM platform.platform_async_installation_activation_receipts WHERE operation_id = in_operation_id AND receipt_kind = 'activation_event') THEN
        RETURN true;
    END IF;
    IF NOT platform.tenancy_activation_step_context_assert_v1(in_operation_id, 'activation.complete', in_tenant_id, in_schema_name) THEN
        RAISE EXCEPTION 'installation activation event is outside its activation transaction' USING ERRCODE = '42501';
    END IF;
    recorded_event := platform.tenancy_activation_digest_uuid_v1('async.activation.event|' || in_operation_id::text);
    payload_sha := platform.tenancy_activation_digest_hex_v1('installation-activation|' || in_operation_id::text || '|' || in_manifest_sha256);
    INSERT INTO platform.outbox (
        event_id, tenant_id, scope_kind, scope_id, authority_group_id, authority_group_generation, authority_domain,
        cutover_generation, writer_origin, envelope_contract_version, event_type, event_version, aggregate_type, aggregate_id,
        aggregate_version, event_sequence, occurred_at, actor_id, correlation_id, causation_id, idempotency_key,
        business_effect_idempotency_key, payload_sha256, payload_schema_sha256, delivery_state, delivery_attempts, lease_generation, next_attempt_at
    ) VALUES (
        recorded_event, in_tenant_id, 'tenant', in_tenant_id, in_operation_id::text, 0, 'installation', 0, 'installation_activation',
        'installation.v1', 'installation.activation.recorded', 1, 'tenancy_onboarding_operation', in_operation_id, 1, 1,
        pg_catalog.clock_timestamp(), in_actor_id, in_operation_id::text, recorded_event, 'activation-event:' || in_operation_id::text,
        'activation-event:' || in_operation_id::text, payload_sha, in_manifest_sha256, 'pending', 0, 0, pg_catalog.clock_timestamp()
    ) ON CONFLICT ON CONSTRAINT outbox_pkey DO NOTHING;
    INSERT INTO platform.jobs (
        id, tenant_id, scope_kind, scope_id, authority_group_id, authority_group_generation, authority_domain, cutover_generation,
        writer_origin, kind, payload_revision, payload_sha256, dedupe_key, effect_class, priority, run_at, state, attempts, max_attempts, lease_generation
    ) VALUES (
        platform.tenancy_activation_digest_uuid_v1('async.activation.job|' || in_operation_id::text), in_tenant_id, 'tenant', in_tenant_id,
        in_operation_id::text, 0, 'installation', 0, 'installation_activation', 'installation.activation.recorded', 1, payload_sha,
        'activation-event:' || in_operation_id::text, 'installation', 0, pg_catalog.clock_timestamp(), 'queued', 0, 8, 0
    ) ON CONFLICT ON CONSTRAINT jobs_pkey DO NOTHING;
    INSERT INTO platform.platform_async_installation_activation_receipts (
        operation_id, receipt_kind, environment_id, tenant_id, event_id, manifest_sha256, created_at
    ) VALUES (in_operation_id, 'activation_event', in_environment_id, in_tenant_id, recorded_event, in_manifest_sha256, pg_catalog.clock_timestamp());
    inserted := true;
    RETURN NOT inserted;
END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_installation_activation_event_v2(uuid, uuid, uuid, uuid, name, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_installation_activation_event_v2(uuid, uuid, uuid, uuid, name, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platform_async_installation_activation_event_v2(uuid, uuid, uuid, uuid, name, text);
DROP TABLE IF EXISTS platform.platform_async_installation_activation_receipts;
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_installation_activation_event_v2(in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_actor_id uuid, in_schema_name name, in_manifest_sha256 text)
RETURNS boolean LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT false; $fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.platform_async_installation_activation_event_v2(uuid, uuid, uuid, uuid, name, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_installation_activation_event_v2(uuid, uuid, uuid, uuid, name, text) TO harpy_api, harpy_worker;
