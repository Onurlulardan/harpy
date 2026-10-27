-- +goose Up
ALTER TABLE platform.tenancy_customer_provisioning_outcomes
ADD COLUMN setup_reissue_operation_id uuid;

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_customer_setup_reissue_target_v1(
    p_tenant_id uuid
)
RETURNS TABLE (
    operation_id uuid, tenant_id uuid, requested_by_subject_id uuid,
    state text, setup_delivery_state text, failure_code text,
    failure_retryable boolean, correlation_id text,
    created_at timestamptz, updated_at timestamptz,
    setup_reissue_operation_id uuid,
    schema_name name, first_user_id uuid, first_admin_email text
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'forgecrud-v2:core-api'
       OR p_tenant_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='customer setup reissue target denied';
    END IF;
    RETURN QUERY
    SELECT operation.operation_id,operation.tenant_id,
        request.requested_by_subject_id,operation.state,
        outcome.setup_delivery_state,outcome.failure_code,
        outcome.failure_retryable,outcome.correlation_id::text,
        operation.created_at,operation.updated_at,
        outcome.setup_reissue_operation_id,operation.schema_name,
        operation.first_user_id,request.first_admin_email
    FROM platform.tenancy_onboarding_operations AS operation
    JOIN platform.tenancy_customer_provisioning_requests AS request
      ON request.operation_id=operation.operation_id
    JOIN platform.tenancy_customer_provisioning_outcomes AS outcome
      ON outcome.operation_id=operation.operation_id
    WHERE operation.tenant_id=p_tenant_id
      AND operation.operation_kind='saas_customer_tenant_provision'
      AND operation.deployment_mode='saas' AND operation.state='succeeded'
      AND outcome.setup_delivery_state IN (
          'permanent_failure','outcome_unknown','reissued'
      );
END;
$function$;
CREATE FUNCTION platform.tenancy_customer_setup_reissue_v1(
    p_operation_id uuid, p_reissue_operation_id uuid, p_tenant_id uuid,
    p_old_artifact_id uuid, p_new_artifact_id uuid
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_outcome platform.tenancy_customer_provisioning_outcomes%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'forgecrud-v2:core-api'
       OR p_operation_id IS NULL OR p_reissue_operation_id IS NULL
       OR p_tenant_id IS NULL OR p_old_artifact_id IS NULL
       OR p_new_artifact_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='customer setup reissue denied';
    END IF;
    SELECT outcome.* INTO v_outcome
    FROM platform.tenancy_customer_provisioning_outcomes AS outcome
    JOIN platform.tenancy_onboarding_operations AS operation
      ON operation.operation_id=outcome.operation_id
    WHERE operation.operation_id=p_operation_id
      AND operation.tenant_id=p_tenant_id
      AND operation.operation_kind='saas_customer_tenant_provision'
      AND operation.deployment_mode='saas' AND operation.state='succeeded'
    FOR UPDATE OF outcome;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='customer setup reissue target conflicts';
    END IF;
    IF v_outcome.setup_delivery_state='reissued'
       AND v_outcome.setup_artifact_id=p_new_artifact_id
       AND v_outcome.setup_reissue_operation_id=p_reissue_operation_id THEN
        RETURN true;
    END IF;
    IF p_old_artifact_id=p_new_artifact_id THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='customer setup reissue artifact conflicts';
    END IF;
    IF v_outcome.setup_delivery_state NOT IN (
        'permanent_failure','outcome_unknown'
    ) OR v_outcome.setup_artifact_id IS DISTINCT FROM p_old_artifact_id THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='customer setup reissue state conflicts';
    END IF;
    UPDATE platform.tenancy_customer_provisioning_outcomes
    SET setup_delivery_state='reissued',setup_artifact_id=p_new_artifact_id,
        setup_reissue_operation_id=p_reissue_operation_id,
        revision=revision+1,updated_at=pg_catalog.statement_timestamp()
    WHERE operation_id=p_operation_id;
    RETURN false;
END;
$function$;
CREATE FUNCTION platform.tenancy_customer_setup_reissue_assert_v1(
    p_operation_id uuid, p_reissue_operation_id uuid, p_tenant_id uuid,
    p_artifact_id uuid, p_schema_name name
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'forgecrud-v2:core-api' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='customer setup reissue assertion denied';
    END IF;
    RETURN EXISTS (
        SELECT 1
        FROM platform.tenancy_customer_provisioning_outcomes AS outcome
        JOIN platform.tenancy_onboarding_operations AS operation
          ON operation.operation_id=outcome.operation_id
        WHERE operation.operation_id=p_operation_id
          AND operation.tenant_id=p_tenant_id
          AND operation.operation_kind='saas_customer_tenant_provision'
          AND operation.deployment_mode='saas' AND operation.state='succeeded'
          AND operation.schema_name=p_schema_name
          AND outcome.setup_delivery_state='reissued'
          AND outcome.setup_artifact_id=p_artifact_id
          AND outcome.setup_reissue_operation_id=p_reissue_operation_id
    );
END;
$function$;
REVOKE ALL ON FUNCTION platform.tenancy_customer_setup_reissue_target_v1(uuid)
FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.tenancy_customer_setup_reissue_v1(
    uuid,uuid,uuid,uuid,uuid
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.tenancy_customer_setup_reissue_assert_v1(
    uuid,uuid,uuid,uuid,name
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_customer_setup_reissue_target_v1(uuid)
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.tenancy_customer_setup_reissue_v1(
    uuid,uuid,uuid,uuid,uuid
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.tenancy_customer_setup_reissue_assert_v1(
    uuid,uuid,uuid,uuid,name
) TO forgecrud_runtime;
-- +goose StatementEnd
