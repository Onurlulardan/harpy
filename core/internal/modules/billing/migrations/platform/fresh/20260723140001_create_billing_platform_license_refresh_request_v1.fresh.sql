-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_platform_license_refresh_request_v1(
    p_environment_id uuid,
    p_tenant_id uuid,
    p_actor_id uuid,
    p_idempotency_key text,
    p_request_sha256 text
)
RETURNS TABLE (
    operation_id uuid,
    replayed boolean,
    created_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.clock_timestamp();
    v_existing platform.billing_platform_license_refresh_requests%ROWTYPE;
    v_head platform.billing_license_heads%ROWTYPE;
    v_snapshot platform.billing_license_snapshots%ROWTYPE;
    v_target platform.billing_license_refresh_targets%ROWTYPE;
    v_operation_id uuid;
    v_operation_state text;
    v_count integer;
    v_target_created boolean;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'forgecrud-v2:core-api' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing platform license refresh denied';
    END IF;
    IF p_environment_id IS NULL
       OR pg_catalog.uuid_extract_version(p_environment_id) <> 7
       OR p_tenant_id IS NULL
       OR p_tenant_id = '00000000-0000-0000-0000-000000000000'::uuid
       OR p_actor_id IS NULL
       OR p_actor_id = '00000000-0000-0000-0000-000000000000'::uuid
       OR p_idempotency_key !~ '^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$'
       OR p_request_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing platform license refresh request is invalid';
    END IF;
    IF NOT EXISTS (
        SELECT 1
        FROM platform.tenancy_deployment_policy AS policy
        JOIN platform.tenancy_tenant_bindings AS binding
          ON binding.environment_id = policy.environment_id
         AND binding.tenant_id = p_tenant_id
         AND binding.purpose = 'onpremise_primary'
        WHERE policy.singleton
          AND policy.environment_id = p_environment_id
          AND policy.deployment_mode = 'onpremise'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing platform license refresh requires onpremise primary';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended(p_environment_id::text, 0)
    );
    SELECT request.* INTO v_existing
    FROM platform.billing_platform_license_refresh_requests AS request
    WHERE request.environment_id = p_environment_id
      AND request.actor_id = p_actor_id
      AND request.idempotency_key = p_idempotency_key
    FOR SHARE;
    IF FOUND THEN
        IF v_existing.tenant_id IS DISTINCT FROM p_tenant_id
           OR v_existing.request_sha256 IS DISTINCT FROM p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'billing platform license idempotency key was reused';
        END IF;
        RETURN QUERY SELECT v_existing.operation_id, true, v_existing.created_at;
        RETURN;
    END IF;
    SELECT pg_catalog.count(*) INTO v_count
    FROM platform.billing_license_heads AS head
    JOIN platform.billing_license_snapshots AS snapshot
      ON snapshot.snapshot_id = head.snapshot_id
    WHERE head.environment_id = p_environment_id;
    IF v_count = 0 THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'billing platform license was not found';
    ELSIF v_count <> 1 THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing platform license authority is ambiguous';
    END IF;
    SELECT head.* INTO v_head
    FROM platform.billing_license_heads AS head
    WHERE head.environment_id = p_environment_id
    FOR SHARE;
    SELECT snapshot.* INTO v_snapshot
    FROM platform.billing_license_snapshots AS snapshot
    WHERE snapshot.snapshot_id = v_head.snapshot_id
    FOR SHARE;
    v_operation_id := pg_catalog.uuidv7();
    INSERT INTO platform.billing_license_refresh_targets (
        license_id, environment_id, licensee_organization_id, delivery_kind,
        observed_sequence, observed_sha256, operation_id,
        next_attempt_at, lease_until, revision, updated_at
    ) VALUES (
        v_head.license_id, v_head.environment_id,
        v_snapshot.licensee_organization_id, v_snapshot.delivery_kind,
        v_head.sequence, v_head.payload_sha256, v_operation_id,
        v_now, NULL, 1, v_now
    ) ON CONFLICT ON CONSTRAINT billing_license_refresh_targets_pkey
      DO NOTHING;
    v_target_created := FOUND;
    SELECT target.* INTO v_target
    FROM platform.billing_license_refresh_targets AS target
    WHERE target.license_id = v_head.license_id
      AND target.environment_id = v_head.environment_id
    FOR UPDATE;
    SELECT refresh.state INTO v_operation_state
    FROM platform.billing_license_refresh_operations AS refresh
    WHERE refresh.operation_id = v_target.operation_id;
    IF NOT v_target_created AND (
        v_target.lease_until > v_now
        OR v_operation_state IN ('pending', 'running')
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55P03',
            MESSAGE = 'billing platform license refresh is in progress';
    END IF;
    IF NOT v_target_created AND v_operation_state = 'fenced' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing platform license refresh target is fenced';
    END IF;
    IF NOT v_target_created THEN
        UPDATE platform.billing_license_refresh_targets AS target
        SET licensee_organization_id = v_snapshot.licensee_organization_id,
            delivery_kind = v_snapshot.delivery_kind,
            observed_sequence = v_head.sequence,
            observed_sha256 = v_head.payload_sha256,
            operation_id = v_operation_id,
            next_attempt_at = v_now,
            lease_until = NULL,
            revision = target.revision + 1,
            updated_at = v_now
        WHERE target.license_id = v_head.license_id
          AND target.environment_id = v_head.environment_id;
    END IF;
    INSERT INTO platform.billing_platform_license_refresh_requests (
        operation_id, license_id, environment_id, tenant_id, actor_id,
        idempotency_key, request_sha256, created_at
    ) VALUES (
        v_operation_id, v_head.license_id, p_environment_id, p_tenant_id,
        p_actor_id, p_idempotency_key, p_request_sha256, v_now
    );
    RETURN QUERY SELECT v_operation_id, false, v_now;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_platform_license_refresh_request_v1(
    uuid, uuid, uuid, text, text
) FROM PUBLIC;
