-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_customer_provision_read_v1(
    p_operation_id uuid, p_requested_by_subject_id uuid
)
RETURNS TABLE (
    operation_id uuid, tenant_id uuid, requested_by_subject_id uuid,
    state text, setup_delivery_state text, failure_code text,
    failure_retryable boolean, correlation_id text,
    created_at timestamptz, updated_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR p_operation_id IS NULL OR p_requested_by_subject_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='customer provisioning read is invalid';
    END IF;
    PERFORM 1 FROM platform.tenancy_deployment_policy AS policy
    WHERE policy.deployment_mode='saas' FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='customer provisioning mode is denied';
    END IF;
    RETURN QUERY SELECT operation.operation_id,operation.tenant_id,
        request.requested_by_subject_id,operation.state,outcome.setup_delivery_state,
        outcome.failure_code,outcome.failure_retryable,outcome.correlation_id::text,
        operation.created_at,operation.updated_at
    FROM platform.tenancy_onboarding_operations AS operation
    JOIN platform.tenancy_customer_provisioning_requests AS request
      ON request.operation_id=operation.operation_id
    JOIN platform.tenancy_customer_provisioning_outcomes AS outcome
      ON outcome.operation_id=operation.operation_id
    WHERE operation.operation_id=p_operation_id
      AND operation.operation_kind='saas_customer_tenant_provision'
      AND request.requested_by_subject_id=p_requested_by_subject_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_customer_provision_read_v1(uuid,uuid)
FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_customer_provision_read_v1(uuid,uuid)
TO forgecrud_runtime;
