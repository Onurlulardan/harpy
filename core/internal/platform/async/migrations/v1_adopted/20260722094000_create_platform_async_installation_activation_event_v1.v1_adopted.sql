-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_installation_activation_event_v1(
    p_operation_id uuid,
    p_environment_id uuid,
    p_tenant_id uuid,
    p_actor_id uuid,
    p_schema_name name,
    p_manifest_sha256 text
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_payload bytea;
    v_payload_sha256 text;
    v_payload_schema_sha256 text;
    v_occurred_at timestamptz := pg_catalog.statement_timestamp();
    v_inserted integer;
    v_existing platform.outbox%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'forgecrud-v2:core-operator-initialize'
       OR p_operation_id IS NULL OR p_environment_id IS NULL
       OR p_tenant_id IS NULL OR p_actor_id IS NULL OR p_schema_name IS NULL
       OR p_manifest_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'installation activation event denied';
    END IF;
    IF NOT platform.tenancy_activation_assert_v1(
        p_operation_id, p_tenant_id, p_schema_name
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'installation activation event conflicts';
    END IF;
    v_payload := pg_catalog.convert_to(pg_catalog.jsonb_build_object(
        'environment_id', p_environment_id::text,
        'manifest_sha256', p_manifest_sha256,
        'operation_id', p_operation_id::text,
        'tenant_id', p_tenant_id::text
    )::text, 'UTF8');
    v_payload_sha256 := pg_catalog.encode(pg_catalog.sha256(v_payload), 'hex');
    v_payload_schema_sha256 := pg_catalog.encode(pg_catalog.sha256(
        pg_catalog.convert_to('installation.activation.v1', 'UTF8')
    ), 'hex');
    INSERT INTO platform.outbox (
        event_id, tenant_id, scope_kind, scope_id,
        authority_group_id, authority_group_generation,
        authority_domain, cutover_generation, writer_origin,
        envelope_contract_version, event_type, event_version,
        aggregate_type, aggregate_id, aggregate_version, event_sequence,
        occurred_at, actor_id, correlation_id, causation_id,
        idempotency_key, business_effect_idempotency_key,
        payload_canonical, payload_sha256, payload_schema_sha256,
        delivery_state, delivery_attempts, lease_generation, next_attempt_at
    ) VALUES (
        p_operation_id, p_tenant_id, 'tenant', p_tenant_id,
        'iam_system_resources', 1, 'iam', 1, 'v2',
        'integration-event.v1', 'installation.activated', 1,
        'installation.environment', p_tenant_id, 1, 1,
        v_occurred_at, p_actor_id,
        'installation:' || p_operation_id::text, p_operation_id,
        'installation.activation:' || p_operation_id::text,
        'installation.activation:' || p_operation_id::text,
        v_payload, v_payload_sha256, v_payload_schema_sha256,
        'pending', 0, 0, v_occurred_at
    ) ON CONFLICT (event_id) DO NOTHING;
    GET DIAGNOSTICS v_inserted = ROW_COUNT;
    IF v_inserted = 1 THEN
        RETURN false;
    END IF;
    SELECT * INTO v_existing
    FROM platform.outbox AS event
    WHERE event.event_id = p_operation_id
    FOR SHARE;
    IF NOT FOUND OR ROW(v_existing.tenant_id, v_existing.scope_kind,
        v_existing.scope_id, v_existing.authority_group_id,
        v_existing.authority_group_generation, v_existing.authority_domain,
        v_existing.cutover_generation, v_existing.writer_origin,
        v_existing.envelope_contract_version, v_existing.event_type,
        v_existing.event_version, v_existing.aggregate_type,
        v_existing.aggregate_id, v_existing.aggregate_version,
        v_existing.event_sequence, v_existing.actor_id,
        v_existing.causation_id, v_existing.idempotency_key,
        v_existing.business_effect_idempotency_key,
        v_existing.payload_canonical, v_existing.payload_sha256,
        v_existing.payload_schema_sha256)
       IS DISTINCT FROM ROW(p_tenant_id, 'tenant', p_tenant_id,
        'iam_system_resources', 1::bigint, 'iam', 1::bigint, 'v2',
        'integration-event.v1', 'installation.activated', 1,
        'installation.environment', p_tenant_id, 1::bigint, 1::bigint,
        p_actor_id, p_operation_id,
        'installation.activation:' || p_operation_id::text,
        'installation.activation:' || p_operation_id::text,
        v_payload, v_payload_sha256, v_payload_schema_sha256) THEN
        RAISE EXCEPTION USING ERRCODE = '23505',
            MESSAGE = 'installation activation event replay conflicts';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_installation_activation_event_v1(
    uuid,uuid,uuid,uuid,name,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_installation_activation_event_v1(
    uuid,uuid,uuid,uuid,name,text
) TO harpy_runtime;
