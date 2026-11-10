-- +goose Up
UPDATE platform.outbox
SET event_type = 'installation.activated',
    event_version = 1,
    authority_domain = 'iam',
    payload_schema_sha256 = '4a44e2f3f9e2669280d713512812010d50bf9f077f4ee1a0709e80c3633dcb5e',
    delivery_state = 'delivered',
    delivered_at = pg_catalog.clock_timestamp()
WHERE event_type = 'installation.activation.recorded'
  AND delivery_state IN ('pending', 'publishing');

UPDATE platform.jobs
SET state = 'succeeded',
    completed_at = pg_catalog.clock_timestamp()
WHERE state IN ('queued', 'running')
  AND kind IN ('installation.activation.recorded', 'iam.initial_admin_setup.deliver');

DROP FUNCTION platform.platform_async_installation_activation_event_v2(uuid, uuid, uuid, uuid, name, text);
DROP FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(uuid, uuid, uuid, uuid, text, name, text);

-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_installation_activation_event_v2(
    in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_actor_id uuid, in_schema_name name, in_manifest_sha256 text
) RETURNS boolean LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
DECLARE operation platform.tenancy_onboarding_operations%ROWTYPE; recorded_event uuid; payload_sha text;
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
        business_effect_idempotency_key, payload_sha256, payload_schema_sha256, delivery_state, delivery_attempts, lease_generation, next_attempt_at, delivered_at
    ) VALUES (
        recorded_event, in_tenant_id, 'tenant', in_tenant_id, in_operation_id::text, 0, 'iam', 0, 'installation_activation',
        'installation.v1', 'installation.activated', 1, 'tenancy_onboarding_operation', in_operation_id, 1, 1,
        pg_catalog.clock_timestamp(), in_actor_id, in_operation_id::text, recorded_event, 'activation-event:' || in_operation_id::text,
        'activation-event:' || in_operation_id::text, payload_sha, '4a44e2f3f9e2669280d713512812010d50bf9f077f4ee1a0709e80c3633dcb5e',
        'delivered', 0, 0, pg_catalog.clock_timestamp(), pg_catalog.clock_timestamp()
    ) ON CONFLICT ON CONSTRAINT outbox_pkey DO NOTHING;
    INSERT INTO platform.jobs (
        id, tenant_id, scope_kind, scope_id, authority_group_id, authority_group_generation, authority_domain, cutover_generation,
        writer_origin, kind, payload_revision, payload_sha256, dedupe_key, effect_class, priority, run_at, state, attempts, max_attempts, lease_generation, completed_at
    ) VALUES (
        platform.tenancy_activation_digest_uuid_v1('async.activation.job|' || in_operation_id::text), in_tenant_id, 'tenant', in_tenant_id,
        in_operation_id::text, 0, 'iam', 0, 'installation_activation', 'installation.activation.recorded', 1, payload_sha,
        'activation-event:' || in_operation_id::text, 'installation', 0, pg_catalog.clock_timestamp(), 'succeeded', 0, 8, 0, pg_catalog.clock_timestamp()
    ) ON CONFLICT ON CONSTRAINT jobs_pkey DO NOTHING;
    INSERT INTO platform.platform_async_installation_activation_receipts (
        operation_id, receipt_kind, environment_id, tenant_id, event_id, manifest_sha256, created_at
    ) VALUES (in_operation_id, 'activation_event', in_environment_id, in_tenant_id, recorded_event, in_manifest_sha256, pg_catalog.clock_timestamp());
    RETURN false;
END;
$fn$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(
    in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_artifact_id uuid, in_artifact_reference text, in_schema_name name, in_manifest_sha256 text
) RETURNS boolean LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
DECLARE operation platform.tenancy_onboarding_operations%ROWTYPE; payload_sha text;
BEGIN
    SELECT * INTO operation FROM platform.tenancy_onboarding_operations WHERE operation_id = in_operation_id;
    IF NOT FOUND OR operation.tenant_id <> in_tenant_id OR operation.environment_id IS DISTINCT FROM in_environment_id
        OR operation.schema_name IS DISTINCT FROM in_schema_name::text OR operation.manifest_sha256 IS DISTINCT FROM in_manifest_sha256 THEN
        RAISE EXCEPTION 'initial admin setup delivery does not match the activation' USING ERRCODE = '42501';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM platform.iam_initial_admin_setup_stages WHERE operation_id = in_operation_id AND tenant_id = in_tenant_id
        AND artifact_id = in_artifact_id AND artifact_reference = in_artifact_reference) THEN
        RAISE EXCEPTION 'initial admin setup delivery has no staged credential' USING ERRCODE = '23514';
    END IF;
    IF EXISTS (SELECT 1 FROM platform.platform_async_installation_activation_receipts WHERE operation_id = in_operation_id AND receipt_kind = 'initial_admin_setup' AND artifact_id = in_artifact_id) THEN
        RETURN true;
    END IF;
    IF NOT platform.tenancy_activation_step_context_assert_v1(in_operation_id, 'activation.complete', in_tenant_id, in_schema_name) THEN
        RAISE EXCEPTION 'initial admin setup delivery is outside its activation transaction' USING ERRCODE = '42501';
    END IF;
    payload_sha := platform.tenancy_activation_digest_hex_v1('initial-admin-setup|' || in_operation_id::text || '|' || in_artifact_id::text);
    INSERT INTO platform.jobs (
        id, tenant_id, scope_kind, scope_id, authority_group_id, authority_group_generation, authority_domain, cutover_generation,
        writer_origin, kind, payload_revision, payload_ref, payload_sha256, dedupe_key, effect_class, priority, run_at, state, attempts, max_attempts, lease_generation, completed_at
    ) VALUES (
        platform.tenancy_activation_digest_uuid_v1('async.admin.job|' || in_artifact_id::text), in_tenant_id, 'tenant', in_tenant_id,
        in_operation_id::text, 0, 'iam', 0, 'installation_activation', 'iam.initial_admin_setup.deliver', 1, in_artifact_reference,
        payload_sha, 'initial-admin:' || in_artifact_id::text, 'installation', 0, pg_catalog.clock_timestamp(), 'succeeded', 0, 8, 0, pg_catalog.clock_timestamp()
    ) ON CONFLICT (id) DO NOTHING;
    INSERT INTO platform.platform_async_installation_activation_receipts (
        operation_id, receipt_kind, environment_id, tenant_id, event_id, artifact_id, manifest_sha256, created_at
    ) VALUES (
        in_operation_id, 'initial_admin_setup', in_environment_id, in_tenant_id,
        platform.tenancy_activation_digest_uuid_v1('async.admin.event|' || in_artifact_id::text), in_artifact_id, in_manifest_sha256, pg_catalog.clock_timestamp()
    );
    RETURN false;
END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_installation_activation_event_v2(uuid, uuid, uuid, uuid, name, text) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(uuid, uuid, uuid, uuid, text, name, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_installation_activation_event_v2(uuid, uuid, uuid, uuid, name, text) TO harpy_api, harpy_worker;
GRANT EXECUTE ON FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(uuid, uuid, uuid, uuid, text, name, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platform_async_installation_activation_event_v2(uuid, uuid, uuid, uuid, name, text);
DROP FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(uuid, uuid, uuid, uuid, text, name, text);
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_installation_activation_event_v2(in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_actor_id uuid, in_schema_name name, in_manifest_sha256 text)
RETURNS boolean LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT false; $fn$;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_artifact_id uuid, in_artifact_reference text, in_schema_name name, in_manifest_sha256 text)
RETURNS boolean LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT false; $fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.platform_async_installation_activation_event_v2(uuid, uuid, uuid, uuid, name, text) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(uuid, uuid, uuid, uuid, text, name, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_installation_activation_event_v2(uuid, uuid, uuid, uuid, name, text) TO harpy_api, harpy_worker;
GRANT EXECUTE ON FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(uuid, uuid, uuid, uuid, text, name, text) TO harpy_api, harpy_worker;
