-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_license_refresh_claim_v1(
    p_operation_id uuid, p_license_id uuid,
    p_licensee_organization_id uuid, p_environment_id uuid,
    p_delivery_kind text, p_component_versions jsonb,
    p_lease_seconds integer
)
RETURNS TABLE (
    operation_id uuid, lease_token uuid, license_id uuid,
    licensee_organization_id uuid, environment_id uuid,
    delivery_kind text, nonce bytea, head_found boolean,
    head_sequence bigint, head_sha256 text,
    database_now timestamptz, lease_expires_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_operation platform.billing_license_refresh_operations%ROWTYPE;
    v_head platform.billing_license_heads%ROWTYPE;
    v_attempt integer;
    v_lease uuid;
    v_now timestamptz := pg_catalog.clock_timestamp();
    v_until timestamptz;
BEGIN
    v_lease := pg_catalog.gen_random_uuid();
    IF pg_catalog.uuid_extract_version(p_operation_id) <> 7
       OR p_license_id IS NULL
       OR p_licensee_organization_id IS NULL
       OR p_environment_id IS NULL
       OR p_delivery_kind NOT IN ('connected', 'airgap', 'development')
       OR p_lease_seconds NOT BETWEEN 5 AND 300
       OR pg_catalog.jsonb_typeof(p_component_versions) <> 'array'
       OR pg_catalog.jsonb_array_length(p_component_versions) NOT BETWEEN 1 AND 64
       OR EXISTS (
           SELECT 1
           FROM pg_catalog.jsonb_array_elements(p_component_versions) AS item
           WHERE pg_catalog.jsonb_typeof(item) <> 'object'
              OR (SELECT pg_catalog.count(*)
                  FROM pg_catalog.jsonb_object_keys(item)) <> 2
              OR NOT item ? 'component' OR NOT item ? 'version'
              OR pg_catalog.octet_length(item ->> 'component') NOT BETWEEN 1 AND 64
              OR pg_catalog.octet_length(item ->> 'version') NOT BETWEEN 1 AND 64
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing license refresh claim is invalid';
    END IF;
    INSERT INTO platform.billing_license_refresh_operations (
        operation_id, license_id, licensee_organization_id, environment_id,
        delivery_kind, nonce, component_versions, state, current_attempt,
        lease_token, lease_until, created_at, updated_at, completed_at
    ) VALUES (
        p_operation_id, p_license_id, p_licensee_organization_id,
        p_environment_id, p_delivery_kind,
        pg_catalog.uuid_send(pg_catalog.gen_random_uuid())
            || pg_catalog.uuid_send(pg_catalog.gen_random_uuid()),
        p_component_versions, 'pending', 0, NULL, NULL, v_now, v_now, NULL
    ) ON CONFLICT ON CONSTRAINT billing_license_refresh_operations_pkey
      DO NOTHING;
    SELECT operation.* INTO v_operation
    FROM platform.billing_license_refresh_operations AS operation
    WHERE operation.operation_id = p_operation_id
    FOR UPDATE;
    IF v_operation.license_id IS DISTINCT FROM p_license_id
       OR v_operation.licensee_organization_id IS DISTINCT FROM
            p_licensee_organization_id
       OR v_operation.environment_id IS DISTINCT FROM p_environment_id
       OR v_operation.delivery_kind IS DISTINCT FROM p_delivery_kind
       OR v_operation.component_versions IS DISTINCT FROM p_component_versions THEN
        RAISE EXCEPTION USING ERRCODE = '23505',
            MESSAGE = 'billing license refresh operation was reused';
    END IF;
    IF v_operation.state = 'fenced' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing license refresh operation is fenced';
    END IF;
    IF v_operation.state = 'running' AND v_operation.lease_until > v_now THEN
        RAISE EXCEPTION USING ERRCODE = '55P03',
            MESSAGE = 'billing license refresh lease is active';
    END IF;
    IF v_operation.state = 'running' THEN
        UPDATE platform.billing_license_refresh_attempts
        SET state = 'lease_expired', finished_at = v_now,
            failure_kind = 'lease_expired'
        WHERE billing_license_refresh_attempts.operation_id = p_operation_id
          AND attempt = v_operation.current_attempt AND state = 'running';
        UPDATE platform.billing_license_refresh_operations
        SET state = 'pending', lease_token = NULL, lease_until = NULL,
            updated_at = v_now
        WHERE billing_license_refresh_operations.operation_id = p_operation_id;
    END IF;
    v_attempt := v_operation.current_attempt + 1;
    v_until := v_now + pg_catalog.make_interval(secs => p_lease_seconds);
    UPDATE platform.billing_license_refresh_operations
    SET state = 'running', current_attempt = v_attempt,
        lease_token = v_lease, lease_until = v_until,
        updated_at = v_now, completed_at = NULL
    WHERE billing_license_refresh_operations.operation_id = p_operation_id
      AND state IN ('pending', 'succeeded');
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'billing license refresh claim raced';
    END IF;
    INSERT INTO platform.billing_license_refresh_attempts (
        operation_id, attempt, lease_token, state,
        claimed_at, lease_until, finished_at, failure_kind
    ) VALUES (
        p_operation_id, v_attempt, v_lease, 'running',
        v_now, v_until, NULL, NULL
    );
    SELECT head.* INTO v_head
    FROM platform.billing_license_heads AS head
    WHERE head.license_id = p_license_id
      AND head.environment_id = p_environment_id;
    RETURN QUERY SELECT p_operation_id, v_lease, p_license_id,
        p_licensee_organization_id, p_environment_id, p_delivery_kind,
        v_operation.nonce, FOUND, COALESCE(v_head.sequence, 0),
        COALESCE(v_head.payload_sha256, ''), v_now, v_until;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_license_refresh_claim_v1(
    uuid, uuid, uuid, uuid, text, jsonb, integer
) FROM PUBLIC;
