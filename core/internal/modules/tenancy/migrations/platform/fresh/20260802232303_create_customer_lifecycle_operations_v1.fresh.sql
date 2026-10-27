-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_customer_lifecycle_accept_v1(
    p_environment_id uuid,p_control_tenant_id uuid,p_actor_id uuid,
    p_target_tenant_id uuid,p_kind text,p_expected_revision bigint,
    p_idempotency_key text,p_request_sha256 bytea,
    p_retention_receipt_id uuid,p_backup_receipt_id uuid,
    p_contributor_receipt_id uuid,p_evidence_sha256 bytea,
    p_proof_verified_at timestamptz,p_now timestamptz
)
RETURNS TABLE (
    outcome text,operation_id uuid,control_tenant_id uuid,target_tenant_id uuid,
    kind text,state text,result_revision bigint,replay boolean,
    created_at timestamptz,updated_at timestamptz,
    completed_at timestamptz,failure_code text
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_existing platform.tenancy_customer_lifecycle_operations%ROWTYPE;
    v_tenant platform.tenants%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_environment_id IS NULL OR p_control_tenant_id IS NULL OR
       p_actor_id IS NULL OR p_target_tenant_id IS NULL OR
       p_target_tenant_id=p_control_tenant_id OR
       p_kind NOT IN ('offboard','purge') OR p_expected_revision < 1 OR
       p_idempotency_key IS NULL OR
       p_idempotency_key !~ '^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$' OR
       p_request_sha256 IS NULL OR pg_catalog.length(p_request_sha256)<>32 OR
       p_now IS NULL OR pg_catalog.abs(EXTRACT(EPOCH FROM
           (pg_catalog.transaction_timestamp()-p_now))) > 300 OR
       (p_kind='offboard' AND (p_retention_receipt_id IS NOT NULL OR
           p_backup_receipt_id IS NOT NULL OR p_contributor_receipt_id IS NOT NULL OR
           p_evidence_sha256 IS NOT NULL OR p_proof_verified_at IS NOT NULL)) OR
       (p_kind='purge' AND (p_retention_receipt_id IS NULL OR
           p_backup_receipt_id IS NULL OR p_contributor_receipt_id IS NULL OR
           p_evidence_sha256 IS NULL OR pg_catalog.length(p_evidence_sha256)<>32 OR
           p_proof_verified_at IS NULL OR p_proof_verified_at>p_now OR
           p_now-p_proof_verified_at>INTERVAL '24 hours')) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Customer tenant lifecycle operation invalid';
    END IF;
    PERFORM 1 FROM platform.tenancy_deployment_policy AS policy
    JOIN platform.tenancy_tenant_bindings AS control_binding
      ON control_binding.environment_id=policy.environment_id AND
         control_binding.tenant_id=p_control_tenant_id AND
         control_binding.purpose='saas_control'
    WHERE policy.singleton AND policy.environment_id=p_environment_id AND
          policy.deployment_mode='saas';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='SaaS control tenant boundary is required';
    END IF;
    SELECT operation.* INTO v_existing
    FROM platform.tenancy_customer_lifecycle_operations AS operation
    WHERE operation.control_tenant_id=p_control_tenant_id AND
          operation.idempotency_key=p_idempotency_key;
    IF FOUND THEN
        IF v_existing.request_sha256<>p_request_sha256 OR
           v_existing.kind<>p_kind OR
           v_existing.target_tenant_id<>p_target_tenant_id OR
           v_existing.expected_revision<>p_expected_revision OR
           v_existing.requested_by_actor_id<>p_actor_id THEN
            RETURN QUERY SELECT 'idempotency_conflict',NULL::uuid,NULL::uuid,
                NULL::uuid,NULL::text,NULL::text,NULL::bigint,NULL::boolean,
                NULL::timestamptz,NULL::timestamptz,NULL::timestamptz,NULL::text;
        ELSE
            RETURN QUERY SELECT 'replay',v_existing.operation_id,
                v_existing.control_tenant_id,v_existing.target_tenant_id,
                v_existing.kind,v_existing.state,v_existing.result_revision,true,
                v_existing.created_at,v_existing.updated_at,
                v_existing.completed_at,v_existing.failure_code;
        END IF;
        RETURN;
    END IF;
    SELECT tenant.* INTO v_tenant
    FROM platform.tenancy_tenant_bindings AS binding
    JOIN platform.tenants AS tenant ON tenant.id=binding.tenant_id
    WHERE binding.environment_id=p_environment_id AND
          binding.tenant_id=p_target_tenant_id AND binding.purpose='saas_customer'
    FOR UPDATE OF tenant;
    IF NOT FOUND THEN
        RETURN QUERY SELECT 'not_found',NULL::uuid,NULL::uuid,NULL::uuid,
            NULL::text,NULL::text,NULL::bigint,NULL::boolean,NULL::timestamptz,
            NULL::timestamptz,NULL::timestamptz,NULL::text;
        RETURN;
    END IF;
    IF v_tenant.lifecycle_revision<>p_expected_revision THEN
        RETURN QUERY SELECT 'revision_conflict',NULL::uuid,NULL::uuid,NULL::uuid,
            NULL::text,NULL::text,NULL::bigint,NULL::boolean,NULL::timestamptz,
            NULL::timestamptz,NULL::timestamptz,NULL::text;
        RETURN;
    END IF;
    IF (p_kind='offboard' AND (v_tenant.deleted_at IS NOT NULL OR
            v_tenant.offboarded_at IS NOT NULL OR v_tenant.purged_at IS NOT NULL)) OR
       (p_kind='purge' AND (v_tenant.deleted_at IS NULL OR
            v_tenant.offboarded_at IS NULL OR v_tenant.purged_at IS NOT NULL)) THEN
        RETURN QUERY SELECT 'state_conflict',NULL::uuid,NULL::uuid,NULL::uuid,
            NULL::text,NULL::text,NULL::bigint,NULL::boolean,NULL::timestamptz,
            NULL::timestamptz,NULL::timestamptz,NULL::text;
        RETURN;
    END IF;
    INSERT INTO platform.tenancy_customer_lifecycle_operations(
        environment_id,control_tenant_id,target_tenant_id,kind,state,
        idempotency_key,request_sha256,expected_revision,result_revision,
        requested_by_actor_id,retention_receipt_id,backup_receipt_id,
        contributor_receipt_id,evidence_sha256,proof_verified_at,
        created_at,updated_at
    ) VALUES (p_environment_id,p_control_tenant_id,p_target_tenant_id,p_kind,
        'REQUESTED',p_idempotency_key,p_request_sha256,p_expected_revision,
        p_expected_revision,p_actor_id,p_retention_receipt_id,p_backup_receipt_id,
        p_contributor_receipt_id,p_evidence_sha256,p_proof_verified_at,p_now,p_now)
    RETURNING * INTO v_existing;
    RETURN QUERY SELECT 'accepted',v_existing.operation_id,
        v_existing.control_tenant_id,v_existing.target_tenant_id,
        v_existing.kind,v_existing.state,v_existing.result_revision,false,
        v_existing.created_at,v_existing.updated_at,
        v_existing.completed_at,v_existing.failure_code;
END;
$function$;

CREATE FUNCTION platform.tenancy_customer_lifecycle_read_v1(
    p_control_tenant_id uuid,p_operation_id uuid
)
RETURNS TABLE (
    operation_id uuid,control_tenant_id uuid,target_tenant_id uuid,
    kind text,state text,result_revision bigint,replay boolean,
    created_at timestamptz,updated_at timestamptz,
    completed_at timestamptz,failure_code text
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_control_tenant_id IS NULL OR p_operation_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Customer tenant lifecycle operation read denied';
    END IF;
    RETURN QUERY SELECT operation.operation_id,operation.control_tenant_id,
        operation.target_tenant_id,operation.kind,operation.state,
        operation.result_revision,false,operation.created_at,operation.updated_at,
        operation.completed_at,operation.failure_code
    FROM platform.tenancy_customer_lifecycle_operations AS operation
    WHERE operation.operation_id=p_operation_id AND
          operation.control_tenant_id=p_control_tenant_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_customer_lifecycle_accept_v1(
    uuid,uuid,uuid,uuid,text,bigint,text,bytea,uuid,uuid,uuid,bytea,
    timestamptz,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.tenancy_customer_lifecycle_read_v1(uuid,uuid) FROM PUBLIC;
