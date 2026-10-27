-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_customer_setup_artifact_bind_v1(
    p_operation_id uuid, p_tenant_id uuid, p_artifact_id uuid
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_existing uuid;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'forgecrud-v2:core-operator-initialize'
       OR p_operation_id IS NULL OR p_tenant_id IS NULL
       OR p_artifact_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='customer setup artifact binding denied';
    END IF;
    PERFORM 1
    FROM platform.tenancy_onboarding_operations AS operation
    JOIN platform.tenancy_customer_provisioning_requests AS request
      ON request.operation_id=operation.operation_id
    WHERE operation.operation_id=p_operation_id
      AND operation.tenant_id=p_tenant_id
      AND operation.operation_kind='saas_customer_tenant_provision'
      AND operation.deployment_mode='saas'
      AND operation.state='finalizing'
    FOR UPDATE OF operation;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='customer setup artifact activation conflicts';
    END IF;
    SELECT outcome.setup_artifact_id INTO v_existing
    FROM platform.tenancy_customer_provisioning_outcomes AS outcome
    WHERE outcome.operation_id=p_operation_id
    FOR UPDATE;
    IF NOT FOUND OR v_existing IS NOT NULL
       AND v_existing IS DISTINCT FROM p_artifact_id THEN
        RAISE EXCEPTION USING ERRCODE='23505',
            MESSAGE='customer setup artifact replay conflicts';
    END IF;
    IF v_existing IS NOT NULL THEN
        RETURN true;
    END IF;
    UPDATE platform.tenancy_customer_provisioning_outcomes
    SET setup_artifact_id=p_artifact_id,revision=revision+1,
        updated_at=pg_catalog.statement_timestamp()
    WHERE operation_id=p_operation_id
      AND setup_delivery_state='pending';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='customer setup delivery state conflicts';
    END IF;
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_customer_setup_artifact_bind_v1(
    uuid,uuid,uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_customer_setup_artifact_bind_v1(
    uuid,uuid,uuid
) TO forgecrud_runtime;
