-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_customer_lifecycle_claim_v1(p_worker_id uuid,p_now timestamptz,p_lease_seconds integer)
RETURNS TABLE (
    operation_id uuid,control_tenant_id uuid,target_tenant_id uuid,
    kind text,state text,result_revision bigint,created_at timestamptz,
    updated_at timestamptz,completed_at timestamptz,failure_code text,
    schema_name text,retention_receipt_id uuid,backup_receipt_id uuid,
    contributor_receipt_id uuid,evidence_sha256 bytea,proof_verified_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_operation_id uuid;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR p_worker_id IS NULL OR
       p_now IS NULL OR p_lease_seconds NOT BETWEEN 30 AND 600 OR
       pg_catalog.abs(EXTRACT(EPOCH FROM
           (pg_catalog.transaction_timestamp()-p_now))) > 300 THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Customer tenant lifecycle claim denied';
    END IF;
    SELECT operation.operation_id INTO v_operation_id
    FROM platform.tenancy_customer_lifecycle_operations AS operation
    WHERE operation.state='REQUESTED' OR
          (operation.state='RUNNING' AND operation.leased_until<p_now)
    ORDER BY operation.created_at,operation.operation_id
    FOR UPDATE SKIP LOCKED LIMIT 1;
    IF NOT FOUND THEN
        RETURN;
    END IF;
    UPDATE platform.tenancy_customer_lifecycle_operations AS operation SET
        state='RUNNING',worker_id=p_worker_id,
        leased_until=p_now+pg_catalog.make_interval(secs=>p_lease_seconds),
        attempt_count=operation.attempt_count+1,updated_at=p_now,
        failure_code=NULL,completed_at=NULL
    WHERE operation.operation_id=v_operation_id;
    RETURN QUERY SELECT operation.operation_id,operation.control_tenant_id,
        operation.target_tenant_id,operation.kind,operation.state,
        operation.result_revision,operation.created_at,operation.updated_at,
        operation.completed_at,operation.failure_code,tenant.schema_name::text,
        operation.retention_receipt_id,operation.backup_receipt_id,
        operation.contributor_receipt_id,operation.evidence_sha256,
        operation.proof_verified_at
    FROM platform.tenancy_customer_lifecycle_operations AS operation
    JOIN platform.tenants AS tenant ON tenant.id=operation.target_tenant_id
    WHERE operation.operation_id=v_operation_id;
END;
$function$;
CREATE FUNCTION platform.tenancy_customer_lifecycle_complete_offboard_v1(p_operation_id uuid,p_worker_id uuid,p_now timestamptz)
RETURNS TABLE (
    operation_id uuid,control_tenant_id uuid,target_tenant_id uuid,
    kind text,state text,result_revision bigint,created_at timestamptz,
    updated_at timestamptz,completed_at timestamptz,failure_code text
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_operation platform.tenancy_customer_lifecycle_operations%ROWTYPE;
    v_revision bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_operation_id IS NULL OR p_worker_id IS NULL OR p_now IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Customer tenant offboard completion denied';
    END IF;
    SELECT operation.* INTO v_operation
    FROM platform.tenancy_customer_lifecycle_operations AS operation
    WHERE operation.operation_id=p_operation_id FOR UPDATE;
    IF NOT FOUND OR v_operation.kind<>'offboard' OR
       v_operation.worker_id IS DISTINCT FROM p_worker_id OR
       v_operation.state NOT IN ('RUNNING','SUCCEEDED') THEN
        RETURN;
    END IF;
    IF v_operation.state='SUCCEEDED' THEN
        RETURN QUERY SELECT v_operation.operation_id,v_operation.control_tenant_id,
            v_operation.target_tenant_id,v_operation.kind,v_operation.state,
            v_operation.result_revision,v_operation.created_at,v_operation.updated_at,
            v_operation.completed_at,v_operation.failure_code;
        RETURN;
    END IF;
    IF v_operation.leased_until<p_now THEN
        RETURN;
    END IF;
    UPDATE platform.tenants AS tenant SET status='SUSPENDED',deleted_at=p_now,
        offboarded_at=p_now,lifecycle_revision=tenant.lifecycle_revision+1,
        updated_at=p_now
    WHERE tenant.id=v_operation.target_tenant_id AND tenant.deleted_at IS NULL AND
          tenant.offboarded_at IS NULL AND tenant.purged_at IS NULL AND
          tenant.lifecycle_revision=v_operation.expected_revision
    RETURNING tenant.lifecycle_revision INTO v_revision;
    IF NOT FOUND THEN
        RETURN;
    END IF;
    UPDATE platform.tenancy_customer_lifecycle_operations AS operation SET
        state='SUCCEEDED',result_revision=v_revision,updated_at=p_now,
        completed_at=p_now,failure_code=NULL
    WHERE operation.operation_id=p_operation_id RETURNING operation.* INTO v_operation;
    RETURN QUERY SELECT v_operation.operation_id,v_operation.control_tenant_id,
        v_operation.target_tenant_id,v_operation.kind,v_operation.state,
        v_operation.result_revision,v_operation.created_at,v_operation.updated_at,
        v_operation.completed_at,v_operation.failure_code;
END;
$function$;
CREATE FUNCTION platform.tenancy_customer_lifecycle_complete_purge_v1(p_operation_id uuid,p_worker_id uuid,p_executor_receipt_id uuid,p_executor_receipt_sha256 bytea,p_now timestamptz)
RETURNS TABLE (
    operation_id uuid,control_tenant_id uuid,target_tenant_id uuid,
    kind text,state text,result_revision bigint,created_at timestamptz,
    updated_at timestamptz,completed_at timestamptz,failure_code text
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_operation platform.tenancy_customer_lifecycle_operations%ROWTYPE;
    v_revision bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_operation_id IS NULL OR p_worker_id IS NULL OR
       p_executor_receipt_id IS NULL OR p_executor_receipt_sha256 IS NULL OR
       pg_catalog.length(p_executor_receipt_sha256)<>32 OR p_now IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Customer tenant purge completion denied';
    END IF;
    SELECT operation.* INTO v_operation
    FROM platform.tenancy_customer_lifecycle_operations AS operation
    WHERE operation.operation_id=p_operation_id FOR UPDATE;
    IF NOT FOUND OR v_operation.kind<>'purge' OR
       v_operation.worker_id IS DISTINCT FROM p_worker_id OR
       v_operation.state NOT IN ('RUNNING','SUCCEEDED') THEN
        RETURN;
    END IF;
    IF v_operation.state='SUCCEEDED' THEN
        IF v_operation.executor_receipt_id IS DISTINCT FROM p_executor_receipt_id OR
           v_operation.executor_receipt_sha256<>p_executor_receipt_sha256 THEN
            RETURN;
        END IF;
        RETURN QUERY SELECT v_operation.operation_id,v_operation.control_tenant_id,
            v_operation.target_tenant_id,v_operation.kind,v_operation.state,
            v_operation.result_revision,v_operation.created_at,v_operation.updated_at,
            v_operation.completed_at,v_operation.failure_code;
        RETURN;
    END IF;
    IF v_operation.leased_until<p_now OR v_operation.proof_verified_at IS NULL OR
       p_now-v_operation.proof_verified_at>INTERVAL '24 hours' THEN
        RETURN;
    END IF;
    UPDATE platform.tenants AS tenant SET purged_at=p_now,
        lifecycle_revision=tenant.lifecycle_revision+1,updated_at=p_now
    WHERE tenant.id=v_operation.target_tenant_id AND tenant.deleted_at IS NOT NULL AND
          tenant.offboarded_at IS NOT NULL AND tenant.purged_at IS NULL AND
          tenant.lifecycle_revision=v_operation.expected_revision
    RETURNING tenant.lifecycle_revision INTO v_revision;
    IF NOT FOUND THEN
        RETURN;
    END IF;
    UPDATE platform.tenancy_customer_lifecycle_operations AS operation SET
        state='SUCCEEDED',result_revision=v_revision,updated_at=p_now,
        completed_at=p_now,executor_receipt_id=p_executor_receipt_id,
        executor_receipt_sha256=p_executor_receipt_sha256,failure_code=NULL
    WHERE operation.operation_id=p_operation_id RETURNING operation.* INTO v_operation;
    RETURN QUERY SELECT v_operation.operation_id,v_operation.control_tenant_id,
        v_operation.target_tenant_id,v_operation.kind,v_operation.state,
        v_operation.result_revision,v_operation.created_at,v_operation.updated_at,
        v_operation.completed_at,v_operation.failure_code;
END;
$function$;
CREATE FUNCTION platform.tenancy_customer_lifecycle_fail_v1(p_operation_id uuid,p_worker_id uuid,p_failure_code text,p_now timestamptz)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_operation_id IS NULL OR p_worker_id IS NULL OR p_failure_code IS NULL OR
       p_failure_code !~ '^[A-Z0-9_]{3,64}$' OR p_now IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Customer tenant lifecycle failure record denied';
    END IF;
    UPDATE platform.tenancy_customer_lifecycle_operations AS operation SET
        state=CASE WHEN p_failure_code='PURGE_EXECUTION_FAILED'
                   THEN 'UNKNOWN' ELSE 'FAILED' END,
        failure_code=p_failure_code,updated_at=p_now,completed_at=p_now
    WHERE operation.operation_id=p_operation_id AND operation.state='RUNNING' AND
          operation.worker_id=p_worker_id AND operation.leased_until>=p_now;
    RETURN FOUND;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_customer_lifecycle_claim_v1(uuid,timestamptz,integer) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.tenancy_customer_lifecycle_complete_offboard_v1(uuid,uuid,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.tenancy_customer_lifecycle_complete_purge_v1(uuid,uuid,uuid,bytea,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.tenancy_customer_lifecycle_fail_v1(uuid,uuid,text,timestamptz) FROM PUBLIC;
